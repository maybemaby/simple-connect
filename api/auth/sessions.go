package auth

import (
	"context"
	"encoding/gob"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/justinas/alice"
)

func init() {
	gob.Register(time.Time{})
}

const AuthLifetime = 30 * 24 * time.Hour
const SessionName = "__s_auth_sess"

func NewSessionManager(secure bool) *scs.SessionManager {
	manager := scs.New()
	manager.Store = memstore.New()

	manager.Lifetime = AuthLifetime
	manager.Cookie.Name = SessionName
	manager.Cookie.HttpOnly = true
	manager.Cookie.SameSite = http.SameSiteLaxMode
	manager.Cookie.Secure = secure
	manager.Cookie.Persist = true
	manager.Cookie.Path = "/"

	return manager
}

func AuthInterceptor(s *scs.SessionManager) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {

		return connect.UnaryFunc(func(ctx context.Context, ar connect.AnyRequest) (connect.AnyResponse, error) {
			res, err := next(ctx, ar)

			res.Header().Add("Vary", "Cookie")

			// var token string

			cookieStr := ar.Header().Get("Cookie")

			if cookieStr == "" {

			}

			return res, err
		})
	}

	return connect.UnaryInterceptorFunc(interceptor)
}

func LoginRpc(r connect.AnyResponse, sessionManager *scs.SessionManager, sessionId string) error {

	return nil
}

func Login(r *http.Request, sessionManager *scs.SessionManager, sessionId string) error {
	err := sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}
	sessionManager.Put(r.Context(), "id", sessionId)
	return nil
}

func Logout(r *http.Request, sessionManager *scs.SessionManager) {
	sessionManager.Destroy(r.Context())
}

func RequireAuthMiddleWare(sessionManager *scs.SessionManager) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			id := sessionManager.GetString(r.Context(), "id")

			if id == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
