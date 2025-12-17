package auth

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

func LoginHandler(w http.ResponseWriter, r *http.Request, authService AuthServiceInterface) {
	verifier := oauth2.GenerateVerifier()
	url := authService.LoginUser(verifier)

	http.SetCookie(w, &http.Cookie{
		Name:     "pkce_verifier",
		HttpOnly: true,
		Value:    verifier,
		Path:     "/",
		MaxAge:   300,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "state",
		HttpOnly: true,
		Value:    verifier,
		Path:     "/",
		MaxAge:   300,
	})

	http.Redirect(w, r, url, http.StatusFound)

}

func RedirectURLHandler(w http.ResponseWriter, r *http.Request, authService AuthServiceInterface) {

	code := r.URL.Query().Get("code")
	cookie, cookieReadErr := r.Cookie("pkce_verifier")
	if cookieReadErr != nil {
		http.Error(w, "Missing PKCE verifier", http.StatusBadRequest)
		return
	}

	token, err := authService.GoogleClient().Exchange(r.Context(), code, oauth2.VerifierOption(cookie.Value))
	if err != nil {
		logger.Error("error on exchanhing the tokem", "Error", err)
		http.Error(w, "Exchange the tokens has failed", http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)

	if !ok {
		logger.Error("No raw token were found")
		http.Error(w, "no raw token exists on the request", http.StatusInternalServerError)
		return
	}

	payload, idValidateErr := idtoken.Validate(r.Context(), rawIDToken, authService.GoogleClient().ClientID)
	if idValidateErr != nil {
		http.Error(w, "Invalid token", http.StatusInternalServerError)
		logger.Error("Error", idValidateErr)
		return
	}

	handlerErr := authService.HandleToken(payload, r.Context())
	if handlerErr != nil {
		logger.Error("error", "error", err)
		return
	}

}
