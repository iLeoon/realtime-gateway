package auth

import (
	"fmt"
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
		http.Error(w, "Exchange the tokens has failed", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)

	if !ok {
		return
	}

	payload, idValidateErr := idtoken.Validate(r.Context(), rawIDToken, authService.GoogleClient().ClientID)
	if idValidateErr != nil {
		logger.Error("Error", idValidateErr)
		return
	}

	authService.HandleToken(payload)

}
