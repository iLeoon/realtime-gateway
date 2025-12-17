package httpserver

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/auth"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/routes"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Start(conf *config.Config, db *pgxpool.Pool) {
	mux := http.NewServeMux()

	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(conf, authRepo)

	mux.Handle("/auth/", routes.AuthRoute(authService))

	logger.Info("Listen to http requests")
	http.ListenAndServe(conf.HttpPort, mux)
}
