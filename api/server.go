package api

import (
	"context"
	"log/slog"
	"net/http"
	"simple-connect/api/auth"
	"simple-connect/gen/proto/api/v1/apiv1connect"

	"github.com/alexedwards/scs/v2"
)

type Server struct {
	mux            *http.ServeMux
	srv            *http.Server
	logger         *slog.Logger
	sessionManager *scs.SessionManager
}

type ServerConfig struct {
	Port     string
	LogLevel slog.Level
}

func NewServer(cfg ServerConfig, isProd bool) (*Server, error) {

	var sessionManager *scs.SessionManager
	var logFormat LoggingFormat

	if isProd {
		logFormat = JSONFormat
		sessionManager = auth.NewSessionManager(true)
	} else {
		logFormat = TEXTFormat
		sessionManager = auth.NewSessionManager(true)
	}

	logger := BootstrapLogger(cfg.LogLevel, logFormat, !isProd)

	return &Server{
		mux: http.NewServeMux(),
		srv: &http.Server{
			Addr: ":" + cfg.Port,
		},
		logger:         logger,
		sessionManager: sessionManager,
	}, nil
}

func (s *Server) MountHandlers() {

	rootMw := RootMiddleware(*s.logger, MiddlewareConfig{
		CorsOrigin:     "*",
		SessionManager: s.sessionManager,
	})

	healthPath, healthHandler := apiv1connect.NewHealthServiceHandler(&HealthService{})
	s.logger.Debug("Mounting health handler at", slog.String("path", healthPath))
	s.mux.Handle(healthPath, rootMw.Then(healthHandler))

}

func (s *Server) Start() error {

	s.MountHandlers()

	s.srv.Handler = s.mux

	s.logger.Info("Starting server at", slog.String("addr", s.srv.Addr))

	return s.srv.ListenAndServe()
}

func (s *Server) Cleanup(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
