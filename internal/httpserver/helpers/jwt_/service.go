package jwt_

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iLeoon/realtime-gateway/internal/config"
)

var (
	invalidTokenErr  = errors.New("invalid token")
	generateTokenErr = errors.New("error on generating the token")
)

type JwtInterface interface {
	GenerateJWT(int) (string, error)
	DecodeJWT(string) (string, error)
}

type jwt_ struct {
	config *config.Config
}

func NewJWTServic(config *config.Config) *jwt_ {
	return &jwt_{
		config: config,
	}
}

func (t_ *jwt_) GenerateJWT(userID int) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    t_.config.JwtIssure,
		Subject:   strconv.Itoa(userID),
	}

	key := []byte(t_.config.JwtSecretKey)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("%w", generateTokenErr)
	}

	return s, nil

}

func (t_ *jwt_) DecodeJWT(token string) (string, error) {

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(t_.config.JwtIssure),
	)

	claims := &jwt.RegisteredClaims{}

	_, err := parser.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(t_.config.JwtSecretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("%w", invalidTokenErr)
	}

	return claims.Subject, nil
}
