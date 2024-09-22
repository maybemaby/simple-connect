package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"simple-connect/api/auth"
	"simple-connect/api/data"
	"simple-connect/api/internal"
	"simple-connect/gen/proto/api/v1/apiv1connect"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Server struct {
	mux            *http.ServeMux
	srv            *http2.Server
	logger         *slog.Logger
	sessionManager *scs.SessionManager
	Addr           string
	allowedHosts   []string
	pool           *pgxpool.Pool
	ctx            context.Context
}

type ServerConfig struct {
	Port         string
	LogLevel     slog.Level
	AllowedHosts []string
}

func NewServer(cfg ServerConfig, isProd bool) (*Server, error) {

	ctx := context.Background()
	var sessionManager *scs.SessionManager
	var logFormat internal.LoggingFormat

	pool, err := data.NewPool(ctx, !isProd)

	if err != nil {
		return nil, err
	}

	if isProd {
		logFormat = internal.JSONFormat
		sessionManager = auth.NewSessionManager(true, pool)
	} else {
		logFormat = internal.TEXTFormat
		sessionManager = auth.NewSessionManager(false, pool)
	}

	logger := internal.BootstrapLogger(cfg.LogLevel, logFormat, !isProd)

	return &Server{
		mux:            http.NewServeMux(),
		srv:            &http2.Server{},
		logger:         logger,
		sessionManager: sessionManager,
		Addr:           fmt.Sprintf(":%s", cfg.Port),
		allowedHosts:   cfg.AllowedHosts,
		pool:           pool,
		ctx:            ctx,
	}, nil
}

func (s *Server) MountHandlers() {

	rootMw := internal.RootMiddleware(*s.logger, internal.MiddlewareConfig{
		CorsOrigin:     s.allowedHosts[0],
		SessionManager: s.sessionManager,
	})

	authMw := rootMw.Append(auth.RequireAuthMiddleWare(s.sessionManager))

	healthPath, healthHandler := apiv1connect.NewHealthServiceHandler(&HealthService{})
	s.logger.Debug("Mounting health handler at", slog.String("path", healthPath))
	s.mux.Handle(healthPath, rootMw.Then(healthHandler))

	authStore := auth.NewAuthService(s.pool)
	authHandler := auth.NewAuthHandler(authStore, s.sessionManager)
	authPath, authRpc := apiv1connect.NewAuthServiceHandler(authHandler)
	s.logger.Debug("Mounting auth handler at", slog.String("path", authPath))
	s.mux.Handle(authPath, rootMw.Then(authRpc))
	s.mux.Handle("POST /auth/login/{$}", rootMw.ThenFunc(authHandler.Login))
	s.mux.Handle("POST /auth/signup/{$}", rootMw.ThenFunc(authHandler.Signup))

	protectedAuthHandler := auth.NewProtectedAuthHandler(authStore, s.sessionManager)
	protectedAuthPath, protectedAuthRpc := apiv1connect.NewProtectedAuthServiceHandler(protectedAuthHandler)
	s.logger.Debug("Mounting protected auth handler at", slog.String("path", protectedAuthPath))
	s.mux.Handle(protectedAuthPath, authMw.Then(protectedAuthRpc))
	s.mux.Handle("POST /auth/logout/{$}", authMw.ThenFunc(protectedAuthHandler.Logout))
}

func (s *Server) Start() error {

	s.MountHandlers()

	s.logger.Info("Starting server at", slog.String("addr", s.Addr))

	return http.ListenAndServe(s.Addr, h2c.NewHandler(s.mux, s.srv))
}

func (s *Server) Cleanup(ctx context.Context) error {
	s.logger.Info("Shutting down server")

	s.pool.Close()
	return nil
}
