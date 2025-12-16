package httpserver

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/auth"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/users"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/routes"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

func Start(conf *config.Config) {
	mux := http.NewServeMux()

	userRepo := users.NewUserRepo("hello world")

	userService := users.NewUserService(userRepo.Database)

	authService := auth.NewAuthService(conf)

	mux.Handle("/user/", routes.UserRoute(userService))
	mux.Handle("/auth/", routes.AuthRoute(authService))

	logger.Info("Listen to http requests")
	http.ListenAndServe(conf.HttpPort, mux)
}
