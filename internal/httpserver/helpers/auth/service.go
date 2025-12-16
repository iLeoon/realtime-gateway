package auth

import (
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/pkg/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

type AuthServiceInterface interface {
	LoginUser(string) string
	GoogleClient() *oauth2.Config
	HandleToken(*idtoken.Payload)
}

type AuthService struct {
	config *config.Config
}

func NewAuthService(config *config.Config) *AuthService {
	return &AuthService{
		config: config,
	}
}

func (as *AuthService) LoginUser(verifier string) string {

	url := as.GoogleClient().AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	return url

}

func (as *AuthService) HandleToken(payload *idtoken.Payload) {

	googleUser := models.GoogleUser{
		Email: payload.Claims["email"].(string),
		Name:  payload.Claims["name"].(string),
	}

	fmt.Println(googleUser)
}

func (as *AuthService) GoogleClient() *oauth2.Config {
	GoogleClient := &oauth2.Config{
		ClientID:     as.config.GoogleClientID,
		ClientSecret: as.config.GoogleClientSecret,
		RedirectURL:  "http://localhost:7000/auth/redirect/oauth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return GoogleClient

}
