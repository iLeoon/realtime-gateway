package httpserver

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/auth"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/jwt_"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/middelware"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/routes"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Start(conf *config.Config, db *pgxpool.Pool, ws http.Handler) {
	mux := http.NewServeMux()
	jwtService := jwt_.NewJWTServic(conf)

	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(conf, authRepo)

	mux.Handle("/auth/", routes.AuthRoute(authService, jwtService, conf))
	mux.Handle("/users/", middelware.AuthGuard(routes.UserRoute(), jwtService))
	mux.Handle("/ws", middelware.AuthGuard(ws, jwtService))

	logger.Info("The websocekt server is up and running..")
	logger.Info("Listening to http requests")
	http.ListenAndServe(conf.HttpPort, mux)
}
