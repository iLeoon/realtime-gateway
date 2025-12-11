package httpserver

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/users"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/routes"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

func Start(conf *config.Config) {
	mux := http.NewServeMux()

	userRepo := users.NewUserRepo("hello world")

	userService := users.NewUserService(userRepo.Database)

	mux.Handle("/user/", routes.UserRoute(userService))

	logger.Info("Listen to http requests")
	http.ListenAndServe(conf.HttpPort, mux)
}
