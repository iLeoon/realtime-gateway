package auth

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/models"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"golang.org/x/oauth2"
)

type Service interface {
	GenerateOAuthURL(verifier string, state string) (url string)
	GoogleClient() *oauth2.Config
	HandleToken(*models.GoogleClaims, context.Context) (*User, *apierror.APIError, int)
	RequiredCookies(*http.Request) (*http.Cookie, *http.Cookie, error)
	FrontChannelError(oauthCode string) (int, *apierror.APIError)
	BackChannelError(error) (int, *apierror.APIError, error)
}

type TokenService interface {
	GenerateHTTPToken(userID string) (httpToken string, err error)
	DecodeGoogleToken(jwtToken string, reqContext context.Context) (*models.GoogleClaims, error)
}

type Handler struct {
	service Service
	token   TokenService
	config  *config.Config
}

func NewHandler(s Service, t TokenService, c *config.Config) *Handler {
	return &Handler{
		service: s,
		token:   t,
		config:  c,
	}
}

// cookieOpts switches between the env variables
func (h *Handler) cookieOpts() (http.SameSite, string, bool) {
	if h.config.IsProduction() {
		return http.SameSiteNoneMode, ".realtimegateway.me", true
	}
	return http.SameSiteLaxMode, "", false
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	authMux := http.NewServeMux()
	authMux.HandleFunc("GET /auth/login", h.Login)
	authMux.HandleFunc("GET /auth/redirect/oauth/google/callback", h.RedirectURL)
	authMux.HandleFunc("POST /auth/logout", h.Logout)
	return authMux

}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sameSite, domain, secure := h.cookieOpts()
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Domain:   domain,
		Value:    "",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	verifier := oauth2.GenerateVerifier()
	stateString := rand.Text()

	url := h.service.GenerateOAuthURL(verifier, stateString)

	sameSite, domian, secure := h.cookieOpts()
	http.SetCookie(w, &http.Cookie{
		Name:     "pkce_verifier",
		Domain:   domian,
		HttpOnly: true,
		Secure:   secure,
		Value:    verifier,
		Path:     "/",
		MaxAge:   0,
		SameSite: sameSite,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "state",
		Domain:   domian,
		HttpOnly: true,
		Secure:   secure,
		Value:    stateString,
		Path:     "/",
		MaxAge:   0,
		SameSite: sameSite,
	})

	http.Redirect(w, r, url, http.StatusFound)

}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	const op errors.Op = "handler.RedirectURL"
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := r.URL.Query().Get("error"); err != "" {
		errorDes := r.URL.Query().Get("error_description")
		statusCode, apiErr := h.service.FrontChannelError(err)
		log.Error.Println("an error occurred while the user trying to authenticate", "error_code", err, "error_description", errorDes)
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		apiErr := apierror.Build(apierror.BadRequestCode,
			"missing required parameters",
			apierror.WithTarget("OAuthParameters"),
		)
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	verifier, stateCookie, err := h.service.RequiredCookies(r)
	if err != nil {
		log.Error.Println("Error", err)
		apiErr := apierror.Build(apierror.BadRequestCode,
			"missing required parameters",
			apierror.WithTarget("OAuthCookies"),
		)
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	if stateCookie.Value != state {
		log.Error.Println("invalid state cookie is being used", "state_cookie", stateCookie.Value, "used_state", state)
		apiErr := apierror.Build(apierror.ForbiddenRequestCode,
			"using invalid parameters",
			apierror.WithTarget("state"),
		)
		apiresponse.Send(w, http.StatusForbidden, apiErr)
		return
	}

	token, err := h.service.GoogleClient().Exchange(ctx, code, oauth2.VerifierOption(verifier.Value))
	if err != nil {
		statusCode, apiErr, backErr := h.service.BackChannelError(err)
		log.Error.Println("unexpected error occurred", backErr)
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Error.Println("expected openid in the configuration scopes")
		apiErr := apierror.Build(apierror.InternalServerErrorCode,
			"authentication has failed", apierror.WithTarget("idtoken"),
			apierror.WithInnerError("MissingOpenIdInTheConfigScopes"))
		apiresponse.Send(w, http.StatusInternalServerError, apiErr)
		return
	}

	payload, err := h.token.DecodeGoogleToken(rawIDToken, r.Context())
	if err != nil {
		var errWrapper = errors.B(path, op, "faild to decode google token", err)
		log.Error.Println(errWrapper)
		var apiErr *apierror.APIError
		var statusCode int
		switch {
		case errors.Is(err, errors.Client):
			apiErr = apierror.Build(apierror.BadRequestCode, "the identity token is invalid",
				apierror.WithTarget("google token"),
				apierror.WithInnerError("InvalidOrExpiredIdentityToken"))
			statusCode = http.StatusBadRequest
		case errors.Is(err, errors.TimeOut):
			apiErr = apierror.TimeOutError("google token")
			statusCode = http.StatusGatewayTimeout

		case errors.Is(err, errors.NotFound):
			apiErr = apierror.Build(apierror.NotFoundRequestCode, "no token was found", apierror.WithTarget("google token"))
			statusCode = http.StatusNotFound

		case errors.Is(err, errors.Internal):
			apiErr = apierror.Build(apierror.NotFoundRequestCode, "unexpected error", apierror.WithTarget("google token"))
			statusCode = http.StatusInternalServerError
		default:
			return
		}
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	user, apiErr, statusCode := h.service.HandleToken(payload, r.Context())
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	jwtToken, err := h.token.GenerateHTTPToken(user.UserID)
	if err != nil {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.FaildToGenerateToken("GeneratingHttpJwtTokenFailed"))
		return
	}

	sameSite, domain, secure := h.cookieOpts()
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Domain:   domain,
		Value:    jwtToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
		MaxAge:   3600,
	})
	http.Redirect(w, r, h.config.Cors+"/chat", http.StatusFound)

}
