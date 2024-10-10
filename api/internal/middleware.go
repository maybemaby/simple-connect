package internal

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
	"github.com/justinas/alice"
	"github.com/unrolled/secure"
)

var RequestIdHeader = "X-Request-Id"

type RequestLoggerContextKey string

const RequestLoggerKey RequestLoggerContextKey = "logger"

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func CorsMiddleware(origin string) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms, X-User-Agent")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequestIdMiddleware() alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			r.Header.Set(RequestIdHeader, uuid.New().String())
			next.ServeHTTP(w, r)
		})
	}
}

func LoggingMiddleware(logger slog.Logger) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{w, http.StatusOK}
			url := r.URL.String()
			requestId := r.Header.Get(RequestIdHeader)
			method := r.Method

			_logger := logger.WithGroup("request").With(slog.String("url", url), slog.String("method", method), slog.String("request_id", requestId))

			_logger.Info("Request received")

			context := context.WithValue(r.Context(), RequestLoggerKey, _logger)

			start := time.Now()
			next.ServeHTTP(rec, r.WithContext(context))

			_logger.Info("Request completed", slog.Duration("duration", time.Since(start)), slog.Int("status", rec.status))
		})
	}
}

func RequestLogger(request *http.Request) *slog.Logger {
	return request.Context().Value(RequestLoggerKey).(*slog.Logger)
}

type MiddlewareConfig struct {
	CorsOrigin     string
	SessionManager *scs.SessionManager
}

func RootMiddleware(logger slog.Logger, cfg MiddlewareConfig) alice.Chain {

	secureMw := secure.New(secure.Options{
		// AllowedHosts:      []string{"http://localhost"},
		HostsProxyHeaders: []string{"X-Forwarded-Host"},
	})

	return alice.New(cfg.SessionManager.LoadAndSave, RequestIdMiddleware(), LoggingMiddleware(logger), CorsMiddleware(cfg.CorsOrigin), secureMw.Handler)
}

func RpcLogger(ctx context.Context) *slog.Logger {
	return ctx.Value(RequestLoggerKey).(*slog.Logger)
}
