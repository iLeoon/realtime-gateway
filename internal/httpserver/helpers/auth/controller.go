package auth

import (
	"crypto/rand"
	"encoding/json"
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/jwt_"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

func LoginHandler(w http.ResponseWriter, r *http.Request, authService AuthServiceInterface) {
	verifier := oauth2.GenerateVerifier()
	stateString := rand.Text()

	url := authService.LoginUser(verifier, stateString)

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

func RedirectURLHandler(w http.ResponseWriter, r *http.Request, authService AuthServiceInterface, jwt jwt_.JwtInterface) {

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		logger.Error("Missing code and state in the request")
		http.Error(w, "Missing required values", http.StatusBadRequest)
		return
	}

	verifier, err := r.Cookie("pkce_verifier")
	if err != nil {
		logger.Error("Missing the verifier cookie in the request.")
		http.Error(w, "Invalid oAuth state", http.StatusBadRequest)
		return
	}

	stateCookie, err := r.Cookie("state")
	if err != nil {
		logger.Error("Missing the state cookie in the request.")
		http.Error(w, "Invalid oAuth state", http.StatusBadRequest)
		return
	}

	if stateCookie.Value != state {
		http.Error(w, "Invalid oAuth state", http.StatusBadRequest)
		return
	}

	token, err := authService.GoogleClient().Exchange(r.Context(), code, oauth2.VerifierOption(verifier.Value))
	if err != nil {
		logger.Error("error on exchanging the token", "Error", err)
		http.Error(w, "Exchange the tokens has failed", http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)

	if !ok {
		logger.Error("No raw token were found")
		http.Error(w, "no raw token exists on the request", http.StatusInternalServerError)
		return
	}

	payload, err := idtoken.Validate(r.Context(), rawIDToken, authService.GoogleClient().ClientID)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusInternalServerError)
		logger.Error("Error", err)
		return
	}

	userID, err := authService.HandleToken(payload, r.Context())
	if err != nil {
		logger.Error("couldn't create or retrive the user", "error", err)
		return
	}

	jwtToken, err := jwt.GenerateJWT(userID)
	if err != nil {
		logger.Error("can't generate jwt token", "Error", err)
		http.Error(w, "error on authorization flow", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
		SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusOK)

	// For testing purposes
	json.NewEncoder(w).Encode(map[string]string{
		"token": jwtToken,
	})

}
