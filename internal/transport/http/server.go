package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/middleware"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/auth"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/conversation"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/token"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/user"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/websocket"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/validation"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler interface {
	Handle()
}

func Start(conf *config.Config, db *pgxpool.Pool, ws http.Handler) {
	rootMux := http.NewServeMux()

	// Initializing the validator
	validator := validator.New(validator.WithRequiredStructEnabled())
	validation.Init(validator)

	// Wraping the api mux with ValidateHeader.
	middleware.ValidateHeaders(rootMux)

	// Wraping the api mux with CORS.
	middleware.Cors(rootMux, conf)

	jwtService := token.NewService(conf)

	authRepo := auth.NewRepo(db)
	authService := auth.NewService(conf, authRepo)
	authHandler := auth.NewHandler(authService, jwtService)

	userRepo := user.NewRepo(db)
	userServ := user.NewService(userRepo)
	userHandler := user.NewHandler(userServ)

	convRepo := conversation.NewRepo(db)
	convServ := conversation.NewService(convRepo)
	convHandler := conversation.NewHandler(convServ)

	wsHandler := websocket.NewHandler(jwtService)

	// Specifying the version.
	rootMux.Handle("/api/v1.0/", http.StripPrefix("/api/v1.0", rootMux))

	userMux := userHandler.RegisterRoutes()
	authMux := authHandler.RegisterRoutes()
	convMux := convHandler.RegisterRoutes()
	wsMux := wsHandler.RegsiterRoutes()

	rootMux.Handle("/auth/", authMux)
	rootMux.Handle("/users/", middleware.AuthGuard(userMux, jwtService))
	rootMux.Handle("/conversations/", middleware.AuthGuard(convMux, jwtService))
	rootMux.Handle("/conversations", middleware.AuthGuard(convMux, jwtService))
	rootMux.Handle("/ws/", middleware.AuthGuard(wsMux, jwtService))

	rootMux.Handle("/ws", middleware.ValidateWsTicket(ws, jwtService))

	logger.Info("The http server is up and running..")
	http.ListenAndServe(conf.HttpPort, rootMux)
}
