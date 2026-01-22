package httpserver

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/auth"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/helpers/token"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/middelware"
	"github.com/iLeoon/realtime-gateway/internal/httpserver/routes"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Start(conf *config.Config, db *pgxpool.Pool, ws http.Handler) {
	rootMux := http.NewServeMux()
	apiMux := http.NewServeMux()

	// wraping the api mux with CORS.
	corsHandler := middelware.Cors(apiMux, conf)

	jwtService := token.NewService(conf)
	authRepo := auth.NewRepo(db)
	authService := auth.NewService(conf, authRepo)

	// specifying the version.
	rootMux.Handle("/api/v1.0/", http.StripPrefix("/api/v1.0", corsHandler))

	apiMux.Handle("/auth/", routes.AuthRoute(authService, jwtService))
	apiMux.Handle("/users/", middelware.AuthGuard(routes.UserRoute(), jwtService))
	apiMux.Handle("/ws", middelware.AuthGuard(ws, jwtService))

	logger.Info("The websocekt server is up and running..")
	logger.Info("Listening to http requests")
	http.ListenAndServe(conf.HttpPort, rootMux)
}
