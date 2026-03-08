package http

import (
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/middleware"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/auth"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/conversation"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/friendrequest"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/message"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/token"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/user"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/websocket"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/validation"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Start(conf *config.Config, db *pgxpool.Pool, ws http.Handler) {
	rootMux := http.NewServeMux()

	// Initializing the validator
	validator := validator.New(validator.WithRequiredStructEnabled())
	validation.Init(validator)

	// Init Rate limiter
	rl := middleware.NewRateLimiter()

	// Wraping the api mux with ValidateHeader.
	handler := middleware.ValidateHeaders(rootMux)

	// Wraping the api mux with CORS.
	handler = middleware.Cors(handler, conf)

	jwtService := token.NewService(conf)

	authRepo := auth.NewRepo(db)
	authService := auth.NewService(conf, authRepo)
	authHandler := auth.NewHandler(authService, jwtService, conf)

	userRepo := user.NewRepo(db)
	userServ := user.NewService(userRepo)
	userHandler := user.NewHandler(userServ)

	msgRepo := message.NewRepo(db)
	msgServ := message.NewService(msgRepo)

	convRepo := conversation.NewRepo(db)
	convServ := conversation.NewService(convRepo)
	convHandler := conversation.NewHandler(convServ, msgServ)

	frRepo := friendrequest.NewRepo(db)
	frServ := friendrequest.NewService(frRepo)
	frHandler := friendrequest.NewHandler(frServ)

	wsHandler := websocket.NewHandler(jwtService)

	userMux := userHandler.RegisterRoutes()
	authMux := authHandler.RegisterRoutes()
	convMux := convHandler.RegisterRoutes()
	frMux := frHandler.RegisterRoutes()
	wsMux := wsHandler.RegsiterRoutes()

	rootMux.Handle("/auth/", middleware.RateLimiter(authMux, rl))
	rootMux.Handle("/users/", middleware.AuthGuard(userMux, jwtService))

	rootMux.Handle("/conversations/", middleware.AuthGuard(convMux, jwtService))
	rootMux.Handle("/conversations", middleware.AuthGuard(convMux, jwtService))

	rootMux.Handle("/friendrequests/", middleware.AuthGuard(frMux, jwtService))
	rootMux.Handle("/friendrequests", middleware.AuthGuard(frMux, jwtService))

	rootMux.Handle("/ws/", middleware.AuthGuard(wsMux, jwtService))
	rootMux.Handle("/ws", middleware.ValidateWsTicket(ws, jwtService))

	// Custom http server configurations
	server := &http.Server{
		Addr:              conf.HttpPort,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          log.NewStdLogger(log.Info),
		Handler:           handler,
	}

	// Register versioned prefix after handler is built so middleware is preserved.
	rootMux.Handle("/api/v1.0/", http.StripPrefix("/api/v1.0", rootMux))

	log.Info.Println("http server is up and running...")

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("coudln't connect to the http server", "error", err)
		os.Exit(1)
	}
}
