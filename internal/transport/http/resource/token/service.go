package token

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/models"
)

const path errors.PathName = "token/service"

var googleHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
}

type service struct {
	config       *config.Config
	parser       *jwt.Parser
	googleParser *jwt.Parser
	signedKey    []byte
	mu           sync.RWMutex
	cachedKeys   map[string]*rsa.PublicKey
	lastFetched  time.Time
}

func NewService(c *config.Config) *service {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(c.JwtIssuer),
	)
	googleParser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer("https://accounts.google.com"),
	)

	return &service{
		config:       c,
		parser:       parser,
		googleParser: googleParser,
		cachedKeys:   make(map[string]*rsa.PublicKey),
		signedKey:    []byte(c.JwtSecretKey),
	}
}

func (s *service) GenerateHTTPToken(userID string) (string, error) {
	return s.EncodeToken(userID, time.Hour*24)
}

func (s *service) GenerateWsToken(userID string) (string, error) {
	return s.EncodeToken(userID, time.Second*60)
}

func (s *service) EncodeToken(userID string, duration time.Duration) (string, error) {
	var op errors.Op = "service.EncodeToken"
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    s.config.JwtIssuer,
		Subject:   userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString(s.signedKey)
	if err != nil {
		return "", errors.B(path, op, errors.Internal, err)
	}

	return jwtToken, nil

}

func (s *service) DecodeToken(jwtToken string) (string, error) {
	var op errors.Op = "service.DecodeToken"
	claims := &jwt.RegisteredClaims{}

	if s.signedKey == nil {
		return "", errors.B(path, op, errors.Internal, "missing singed key")
	}
	_, err := s.parser.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (any, error) {
		return s.signedKey, nil
	})

	if err != nil {
		return "", errors.B(path, op, errors.Client, "invalid token is being used", err)
	}

	if claims.Issuer != s.config.JwtIssuer {
		return "", errors.B(path, op, errors.Client, fmt.Errorf("invalid issuer expected: %v and received %v", s.config.JwtIssuer, claims.Issuer))
	}

	if claims.Subject == "" {
		return "", errors.B(path, op, errors.Client, "missing subject in claims")
	}

	return claims.Subject, nil
}

func (s *service) DecodeGoogleToken(jwtToken string, ctx context.Context) (*models.GoogleClaims, error) {
	const op errors.Op = "service.DecodeGoogleToken"
	claims := &models.GoogleClaims{}

	_, err := s.googleParser.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.B(path, op, errors.Client, "missing kid in header")
		}
		return s.cacheLocked(ctx, kid)
	})

	if err != nil {
		return nil, errors.B(path, op, "invalid token is being used", err)
	}

	if claims.Audience[0] != s.config.GoogleClientID {
		return nil, errors.B(path, op, errors.Client, fmt.Errorf("invalid audience passed: %v", claims.Audience[0]))
	}

	return claims, nil
}

// getGooglePublicKey returns the RSA public key for the given key ID.
func (s *service) getGooglePublicKey(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	const op errors.Op = "service.getGooglePublicKey"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v1/certs", nil)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil, errors.B(path, op, errors.Client, "user canceled the request", err)
		case errors.Is(err, context.DeadlineExceeded):
			return nil, errors.B(path, op, errors.TimeOut, "timeout hit when sending a request to fetch google certs", err)
		default:
			return nil, errors.B(path, op, errors.Internal, "unexpected error while making a request to google client", err)
		}
	}

	//nolint:gosec
	resp, err := googleHTTPClient.Do(req)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil, errors.B(path, op, errors.Client, "user canceled the request", err)
		case errors.Is(err, context.DeadlineExceeded):
			return nil, errors.B(path, op, errors.TimeOut, "timeout hit when fetching google certs", err)
		default:
			return nil, errors.B(path, op, errors.Internal, "unexpected error fetching google certs", err)
		}
	}
	defer resp.Body.Close()

	var certs map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&certs); err != nil {
		return nil, errors.B(path, op, errors.Client, "invalid json format", err)
	}

	parsedKeys := make(map[string]*rsa.PublicKey)
	for kid, pemStr := range certs {
		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pemStr))
		if err != nil {
			continue
		}
		parsedKeys[kid] = key
	}

	return parsedKeys, nil
}

func (s *service) cacheLocked(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	const op errors.Op = "service.cacheLocked"
	s.mu.RLock()
	key, ok := s.cachedKeys[kid]
	s.mu.RUnlock()
	if ok && time.Since(s.lastFetched) < 1*time.Hour {
		return key, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if key, ok = s.cachedKeys[kid]; ok && time.Since(s.lastFetched) < 1*time.Hour {
		return key, nil
	}

	googleKeys, err := s.getGooglePublicKey(ctx)
	if err != nil {
		return nil, errors.B(path, op, err)
	}
	s.cachedKeys = googleKeys
	s.lastFetched = time.Now()

	cachedKey, ok := s.cachedKeys[kid]
	if !ok {
		return nil, errors.B(path, op, errors.NotFound, fmt.Errorf("key %s not found in Google's current list", kid))
	}
	return cachedKey, nil
}
