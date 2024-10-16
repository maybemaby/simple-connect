package auth

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justinas/alice"
)

func init() {
	gob.Register(time.Time{})
}

type SessionData struct {
	UserID string `json:"userId"`
}

const SessionUserKey = "user_id"
const AuthLifetime = 30 * 24 * time.Hour
const SessionName = "__s_auth_sess"

func NewMemorySessionManager(secure bool, domain string) *scs.SessionManager {
	manager := scs.New()
	manager.Store = memstore.New()

	manager.Lifetime = AuthLifetime
	manager.Cookie.Name = SessionName
	manager.Cookie.HttpOnly = true
	manager.Cookie.SameSite = http.SameSiteLaxMode
	manager.Cookie.Secure = secure
	manager.Cookie.Persist = true
	manager.Cookie.Path = "/"

	if domain != "" {
		manager.Cookie.Domain = domain
	}

	return manager
}

func NewSessionManager(secure bool, domain string, pool *pgxpool.Pool) *scs.SessionManager {
	manager := scs.New()
	manager.Store = pgxstore.New(pool)

	manager.Lifetime = AuthLifetime
	manager.Cookie.Name = SessionName
	manager.Cookie.HttpOnly = true
	manager.Cookie.SameSite = http.SameSiteLaxMode
	manager.Cookie.Secure = secure
	manager.Cookie.Persist = true
	manager.Cookie.Path = "/"

	if domain != "" {
		manager.Cookie.Domain = domain
	}

	return manager
}

func Login(r *http.Request, sessionManager *scs.SessionManager, data SessionData) error {
	err := sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}
	sessionManager.Put(r.Context(), SessionUserKey, data.UserID)
	return nil
}

func Logout(r *http.Request, sessionManager *scs.SessionManager) {
	sessionManager.Destroy(r.Context())
}

func RequireAuthMiddleWare(sessionManager *scs.SessionManager) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userId := sessionManager.GetString(r.Context(), SessionUserKey)

			if userId == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
