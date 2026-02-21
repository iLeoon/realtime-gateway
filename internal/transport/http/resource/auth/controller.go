package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apiresponse"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

type Service interface {
	GenerateOAuthUrl(verifier string, state string) (url string)
	GoogleClient() (config *oauth2.Config)
	HandleToken(p *idtoken.Payload, ctx context.Context) (u *User, a *apierror.APIError, statusCode int)
	RequiredCookies(r *http.Request) (verifier, state *http.Cookie, err error)
	FrontChannelError(oauthCode string) (statusCode int, a *apierror.APIError)
	BackChannelError(err error) (statusCode int, a *apierror.APIError, error error)
	TestLoadService(ctx context.Context, testUser ProviderIdentity) (*User, error)
}

type TokenService interface {
	GenerateHttpToken(userId string) (httpToken string, err error)
}

type Handler struct {
	service Service
	token   TokenService
}

func NewHandler(s Service, t TokenService) *Handler {
	return &Handler{
		service: s,
		token:   t,
	}
}

func (h *Handler) RegisterRoutes() *http.ServeMux {
	authMux := http.NewServeMux()
	authMux.HandleFunc("GET /auth/login", h.Login)
	authMux.HandleFunc("GET /auth/redirect/oauth/google/callback", h.RedirectURL)
	authMux.HandleFunc("POST /auth/test", h.TestDBLoad)
	return authMux

}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	verifier := oauth2.GenerateVerifier()
	stateString := rand.Text()

	url := h.service.GenerateOAuthUrl(verifier, stateString)

	http.SetCookie(w, &http.Cookie{
		Name:     "pkce_verifier",
		HttpOnly: true,
		Value:    verifier,
		Path:     "/",
		MaxAge:   0,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "state",
		HttpOnly: true,
		Value:    stateString,
		Path:     "/",
		MaxAge:   0,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, url, http.StatusFound)

}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {

	if err := r.URL.Query().Get("error"); err != "" {
		errorDes := r.URL.Query().Get("error_description")
		statusCode, apiErr := h.service.FrontChannelError(err)
		log.Error.Println("an error occured while the user trying to authenticate", "error_code", err, "error_description", errorDes)
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

	token, err := h.service.GoogleClient().Exchange(r.Context(), code, oauth2.VerifierOption(verifier.Value))
	if err != nil {
		statusCode, apiErr, err := h.service.BackChannelError(err)
		log.Error.Println("unexpected error occurred", "error", err)
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Error.Println("expected openid in the configuration scopes")
		apiErr := apierror.Build(apierror.InternalServerErrorCode,
			"authentication has failed", apierror.WithTarget("idtoken"),
			apierror.WithInnerError("MissingOpenIdInTheConfigScopes"))
		apiresponse.Send(w, http.StatusInternalServerError, apiErr)
		return
	}

	payload, err := idtoken.Validate(r.Context(), rawIdToken, h.service.GoogleClient().ClientID)
	if err != nil {
		log.Error.Println("identity tokne is either valid or expired", err, "idtoken", payload)

		apiErr := apierror.Build(apierror.BadRequestCode, "the identity token is invalid",
			apierror.WithTarget("idtoken"),
			apierror.WithInnerError("InvalidOrExpiredIdentityToken"))
		apiresponse.Send(w, http.StatusBadRequest, apiErr)
		return
	}

	userId, apiErr, statusCode := h.service.HandleToken(payload, r.Context())
	if apiErr != nil {
		apiresponse.Send(w, statusCode, apiErr)
		return
	}

	jwtToken, err := h.token.GenerateHttpToken(userId.UserID)
	if err != nil {
		apiresponse.Send(w, http.StatusInternalServerError, apierror.FaildToGenerateToken("GeneratingHttpJwtTokenFailed"))
		return
	}

	w.Header().Set("Authorization", "Bearer "+jwtToken)
	w.WriteHeader(http.StatusCreated)

	// For testing purposes
	json.NewEncoder(w).Encode(map[string]string{
		"token": jwtToken,
	})
}

func (h *Handler) TestDBLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var provider ProviderIdentity

	err := json.NewDecoder(r.Body).Decode(&provider)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := h.service.TestLoadService(r.Context(), provider)
	fmt.Println(err)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{"userId": user.UserID, "email": user.Email, "name": user.UserName})
	fmt.Println(user)
}
