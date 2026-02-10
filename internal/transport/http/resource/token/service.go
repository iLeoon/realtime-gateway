package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/errors"
)

const path errors.PathName = "token/service"

type service struct {
	config    *config.Config
	parser    *jwt.Parser
	signedKey []byte
}

func NewService(c *config.Config) *service {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(c.JwtIssuer),
	)
	return &service{
		config:    c,
		parser:    parser,
		signedKey: []byte(c.JwtSecretKey),
	}
}

func (s *service) GenerateHttpToken(userId string) (string, error) {
	return s.EncodeToken(userId, time.Hour*24)
}

func (s *service) GenerateWsToken(userId string) (string, error) {
	return s.EncodeToken(userId, time.Second*60)
}

func (s *service) EncodeToken(userId string, duration time.Duration) (string, error) {
	var op errors.Op = "service.EncodeToken"
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    s.config.JwtIssuer,
		Subject:   userId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString(s.signedKey)
	if err != nil {
		return "", errors.B(path, op, errors.Internal, err)
	}

	return jwtToken, nil

}

func (s *service) DecodeToken(jwtToken string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	var op errors.Op = "service.DecodeToken"
	if s.signedKey == nil {
		return "", errors.B(path, op, errors.Internal, "missing singed key")
	}
	_, err := s.parser.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (any, error) {
		return s.signedKey, nil
	})

	if err != nil {
		return "", errors.B(path, op, errors.Client, fmt.Errorf("invalid token: %s", err))
	}

	if claims.Issuer != s.config.JwtIssuer {
		return "", errors.B(path, op, errors.Client, fmt.Errorf("invalid issuer expected: %v and recieved %v", s.config.JwtIssuer, claims.Issuer))
	}

	if claims.Subject == "" {
		return "", errors.B(path, op, errors.Client, "missing subject in claims")
	}

	return claims.Subject, nil
}
