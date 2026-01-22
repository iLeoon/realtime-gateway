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
	InvalidTokenErr  = errors.New("Invalid token")
	GenerateTokenErr = errors.New("Error on generating the token")
)

type Service interface {
	EncodeJwt(userId int) (jwtToken string, err error)
	DecodeJwt(jwtToken string) (userId string, err error)
}

type service struct {
	config *config.Config
}

func NewService(c *config.Config) Service {
	return &service{
		config: c,
	}
}

func (s *service) EncodeJwt(userId int) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    s.config.JwtIssure,
		Subject:   strconv.Itoa(userId),
	}

	key := []byte(s.config.JwtSecretKey)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("%w:%v", GenerateTokenErr, err)
	}

	return jwtToken, nil

}

func (s *service) DecodeJwt(jwtToken string) (string, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(s.config.JwtIssure),
	)

	claims := &jwt.RegisteredClaims{}

	_, err := parser.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (any, error) {
		return []byte(s.config.JwtSecretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("%w:%v", InvalidTokenErr, err)
	}

	return claims.Subject, nil
}
