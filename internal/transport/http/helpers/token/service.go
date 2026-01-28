package token

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/iLeoon/realtime-gateway/internal/config"
	"strconv"
	"time"
)

var (
	InternalAuthErr = errors.New("Internal authentication failed")
	SubjectTokenErr = errors.New("The token subject is missing")
)

const (
	ClientError = "ClientError"
	ServerError = "ServerError"
)

type ErrorWrapper struct {
	Category string
	Message  string
	Err      error
}

// Implement the error interface
func (e *ErrorWrapper) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Category, e.Message, e.Err)
}

type Service interface {
	EncodeToken(userId int, duration time.Duration) (jwtToken string, err error)
	DecodeToken(jwtToken string) (userId string, err *ErrorWrapper)
	GenerateHttpToken(userId int) (httpToken string, err error)
	GenerateWsToken(userId int) (wsToken string, err error)
}

type service struct {
	config    *config.Config
	parser    *jwt.Parser
	signedKey []byte
}

func NewService(c *config.Config) Service {
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

func (s *service) GenerateHttpToken(userId int) (string, error) {
	return s.EncodeToken(userId, time.Hour*24)
}

func (s *service) GenerateWsToken(userId int) (string, error) {
	return s.EncodeToken(userId, time.Second*60)
}

func (s *service) EncodeToken(userId int, duration time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    s.config.JwtIssuer,
		Subject:   strconv.Itoa(userId),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString(s.signedKey)
	if err != nil {
		return "", err
	}

	return jwtToken, nil

}

func (s *service) DecodeToken(jwtToken string) (string, *ErrorWrapper) {
	claims := &jwt.RegisteredClaims{}
	_, err := s.parser.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (any, error) {
		if s.signedKey == nil {
			return "", InternalAuthErr
		}
		return s.signedKey, nil
	})

	if err != nil {
		if errors.Is(err, InternalAuthErr) {
			return "", &ErrorWrapper{
				Category: ServerError,
				Message:  "Internal configuration is wrong or missing",
				Err:      err,
			}
		}
		return "", &ErrorWrapper{
			Category: ClientError,
			Message:  "Invalid or expired token",
			Err:      err,
		}
	}

	if claims.Issuer != s.config.JwtIssuer {
		return "", &ErrorWrapper{
			Category: ClientError,
			Message:  "The token issuer is invalid",
			Err:      fmt.Errorf("expected %s, got %v", s.config.JwtIssuer, claims.Issuer),
		}
	}

	if claims.Subject == "" {
		return "", &ErrorWrapper{
			Category: ClientError,
			Message:  "The token subject is missing",
			Err:      SubjectTokenErr,
		}
	}

	return claims.Subject, nil
}
