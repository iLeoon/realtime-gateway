package auth

import (
	"context"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

type Service interface {
	GenerateOAuthUrl(verifier string, state string) (url string)
	GoogleClient() (config *oauth2.Config)
	HandleToken(*idtoken.Payload, context.Context) (userId int, err error)
}

type service struct {
	config *config.Config
	repo   Repository
}

func NewService(c *config.Config, r Repository) Service {
	return &service{
		config: c,
		repo:   r,
	}
}

func (s *service) GenerateOAuthUrl(verifier string, state string) string {
	url := s.GoogleClient().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	return url

}

func (s *service) HandleToken(p *idtoken.Payload, ctx context.Context) (int, error) {
	User := ProviderIdentity{
		ProviderID: p.Subject,
		Email:      p.Claims["email"].(string),
		Name:       p.Claims["name"].(string),
		Provider:   "google",
	}
	userId, err := s.repo.CreateOrUpdateUser(ctx, User)

	if err != nil {
		return 0, err
	}
	return userId, err
}

func (s *service) GoogleClient() *oauth2.Config {
	GoogleConfig := &oauth2.Config{
		ClientID:     s.config.GoogleClientID,
		ClientSecret: s.config.GoogleClientSecret,
		RedirectURL:  s.config.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return GoogleConfig

}
