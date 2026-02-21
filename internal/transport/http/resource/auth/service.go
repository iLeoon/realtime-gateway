package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

const path errors.PathName = "auth/service"

type Repository interface {
	CreateOrUpdateUser(ctx context.Context, pi ProviderIdentity) (u *User, err error)
}

type service struct {
	config      *config.Config
	repo        Repository
	oauthConfig *oauth2.Config
	timeout     time.Duration
}

func NewService(c *config.Config, authRepo Repository) *service {
	return &service{
		config: c,
		repo:   authRepo,
		oauthConfig: &oauth2.Config{
			ClientID:     c.GoogleClientID,
			ClientSecret: c.GoogleClientSecret,
			RedirectURL:  c.RedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		timeout: time.Second * 2,
	}
}

func (s *service) GenerateOAuthUrl(verifier string, state string) string {
	url := s.GoogleClient().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	return url

}

func (s *service) HandleToken(p *idtoken.Payload, ctx context.Context) (*User, *apierror.APIError, int) {
	var apiErr *apierror.APIError
	var statusCode int
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	email, ok := p.Claims["email"].(string)
	if !ok {
		apiErr = apierror.Build(apierror.BadRequestCode, "email is missing", apierror.WithTarget("email"))
		statusCode = http.StatusBadRequest
		return nil, apiErr, statusCode
	}

	name, ok := p.Claims["name"].(string)
	if !ok {
		apiErr = apierror.Build(apierror.BadRequestCode, "name is missing", apierror.WithTarget("email"))
		statusCode = http.StatusBadRequest
		return nil, apiErr, statusCode
	}

	provider := ProviderIdentity{
		ProviderID: p.Subject,
		Email:      email,
		Name:       name,
		Provider:   "google",
	}

	user, err := s.repo.CreateOrUpdateUser(ctx, provider)
	if err != nil {
		log.Error.Println("can't retrieve or create user", "error", err)
		apiErr, statusCode = apierror.ErrorMapper(err, "auth")
		return nil, apiErr, statusCode
	}
	return user, nil, 0
}

func (s *service) GoogleClient() *oauth2.Config {
	return s.oauthConfig
}

func (s *service) RequiredCookies(r *http.Request) (verifier, state *http.Cookie, err error) {
	const op errors.Op = "service.RequiredCookies"
	verifier, err = r.Cookie("pkce_verifier")
	if err != nil {
		return nil, nil, errors.B(path, op, errors.Client, fmt.Errorf("missing verifier cookie: %w", err))
	}

	state, err = r.Cookie("state")
	if err != nil {
		return nil, nil, errors.B(path, op, errors.Client, fmt.Errorf("missing state cookie: %w", err))
	}

	return verifier, state, nil
}

// FrontChannelError follows RPC-style error mapping.
// It translates external OAuth2 error codes into internal status codes
// and structured APIErrors for consistent front-channel response handling.
func (s *service) FrontChannelError(oauthCode string) (int, *apierror.APIError) {
	var code apierror.Code
	var statusCode int
	switch oauthCode {
	case "server_error":
		code = apierror.BadGatewayCode
		statusCode = http.StatusBadGateway
	case "temporarily_unavailable":
		code = apierror.ServiceUnavailable
		statusCode = http.StatusServiceUnavailable
	case "invalid_request":
		code = apierror.BadRequestCode
		statusCode = http.StatusBadRequest
	case "unauthorized_client":
		code = apierror.BadRequestCode
		statusCode = http.StatusBadRequest
	case "access_denied":
		code = apierror.ForbiddenRequestCode
		statusCode = http.StatusForbidden
	case "invalid_scope":
		code = apierror.BadRequestCode
		statusCode = http.StatusBadRequest
	case "unsupported_response_type":
		code = apierror.BadRequestCode
		statusCode = http.StatusBadRequest
	default:
		code = apierror.InternalServerErrorCode
		statusCode = http.StatusBadRequest
	}
	return statusCode, apierror.Build(code,
		"an issue occurred while trying to authenticate",
		apierror.WithTarget("oauth"),
		apierror.WithInnerError("IssueWithOAuthFlow"))

}

// BackChannelError follows the RPC pattern for error handling.
// It returns a result string, a structured APIError, and an HTTP status code
// to ensure the caller has the full context of the server-to-server operation.
func (s *service) BackChannelError(err error) (int, *apierror.APIError, error) {
	const op errors.Op = "service.BackChannelError"
	var statusCode int
	apiErr := apierror.Build(apierror.BadRequestCode, "unexpected error occured",
		apierror.WithTarget("token"),
		apierror.WithInnerError("ExchangeTokenFaild"),
	)

	if rErr, ok := err.(*oauth2.RetrieveError); ok {
		statusCode = http.StatusBadRequest
		return statusCode, apiErr, errors.B(path, op, errors.Client, fmt.Errorf("error_code:%s, error_description:%s", rErr.ErrorCode, rErr.ErrorDescription))
	} else {
		statusCode = http.StatusInternalServerError
		return statusCode, apiErr, errors.B(path, op, errors.Internal, err)
	}
}
func (s *service) TestLoadService(rCtx context.Context, testUser ProviderIdentity) (*User, error) {

	ctx, cancel := context.WithTimeout(rCtx, 10*time.Second)
	defer cancel()
	user, err := s.repo.CreateOrUpdateUser(ctx, testUser)

	if err != nil {
		return nil, err
	}
	return user, nil
}
