package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/auth"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/conversation"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/token"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/helpers/user"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/middleware"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/routes"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/validation"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler interface {
	Handle()
}

func Start(conf *config.Config, db *pgxpool.Pool, ws http.Handler) {
	rootMux := http.NewServeMux()
	apiMux := http.NewServeMux()

	var handler http.Handler = apiMux
	// Initializing the validator
	validator := validator.New(validator.WithRequiredStructEnabled())
	validation.Init(validator)

	// Wraping the api mux with ValidateHeader.
	handler = middleware.ValidateHeaders(handler)

	// Wraping the api mux with CORS.
	handler = middleware.Cors(handler, conf)

	jwtService := token.NewService(conf)

	authRepo := auth.NewRepo(db)
	authService := auth.NewService(conf, authRepo)

	userRepo := user.NewRepo(db)
	userServ := user.NewService(userRepo)

	convRepo := conversation.NewRepo(db)
	convServ := conversation.NewService(convRepo)

	// Specifying the version.
	rootMux.Handle("/api/v1.0/", http.StripPrefix("/api/v1.0", handler))

	// Defining handlers
	authHandler := routes.Auth(authService, jwtService)
	usersHandler := middleware.AuthGuard(routes.User(userServ), jwtService)
	conversationHandler := middleware.AuthGuard(routes.Conversation(convServ), jwtService)
	wsHandler := middleware.AuthGuard(routes.Websocket(jwtService), jwtService)

	apiMux.Handle("/auth/", authHandler)
	apiMux.Handle("/users/", usersHandler)

	apiMux.Handle("/conversations/", conversationHandler)
	apiMux.Handle("/conversations", conversationHandler)

	apiMux.Handle("/ws/", wsHandler)
	apiMux.Handle("/ws", middleware.ValidateWsTicket(ws, jwtService))

	logger.Info("The websocekt server is up and running..")
	logger.Info("Listening to http requests")
	http.ListenAndServe(conf.HttpPort, rootMux)
}
