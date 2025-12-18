package jwt_

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iLeoon/realtime-gateway/internal/config"
)

type JwtInterface interface {
	GenerateJWT(int) string
}

type jwt_ struct {
	config *config.Config
}

func NewJWTServic(config *config.Config) *jwt_ {
	return &jwt_{
		config: config,
	}
}

func (t *jwt_) GenerateJWT(userID int) string {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "RealTime-Gateway",
		Subject:   strconv.Itoa(userID),
	}

	key := []byte(t.config.JwtSecretKey)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString(key)

	return s

}
