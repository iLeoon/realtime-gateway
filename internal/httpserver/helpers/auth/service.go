package auth

import (
	"context"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/pkg/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

type AuthServiceInterface interface {
	LoginUser(string, string) string
	GoogleClient() *oauth2.Config
	HandleToken(*idtoken.Payload, context.Context) (int, error)
}

type AuthService struct {
	config *config.Config
	repo   AuthRepositpryInterface
}

func NewAuthService(config *config.Config, repo AuthRepositpryInterface) *AuthService {
	return &AuthService{
		config: config,
		repo:   repo,
	}
}

func (as *AuthService) LoginUser(verifier string, state string) string {
	url := as.GoogleClient().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	return url

}

func (as *AuthService) HandleToken(payload *idtoken.Payload, ctx context.Context) (int, error) {

	User := models.ProviderUser{
		ProviderID: payload.Subject,
		Email:      payload.Claims["email"].(string),
		Name:       payload.Claims["name"].(string),
		Provider:   "google",
	}
	userID, err := as.repo.HandleLogins(ctx, User)

	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (as *AuthService) GoogleClient() *oauth2.Config {
	GoogleClient := &oauth2.Config{
		ClientID:     as.config.GoogleClientID,
		ClientSecret: as.config.GoogleClientSecret,
		RedirectURL:  as.config.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return GoogleClient

}
