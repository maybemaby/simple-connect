package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"simple-connect/api/internal"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const OAUTH_STATE_SESSION_KEY = "oauth_state"
const OAUTH_VERIFIER_SESSION_KEY = "oauth_verifier"

var ErrStateMismatch = errors.New("state mismatch")
var googleOAuthConfig *oauth2.Config

type ProviderHandler struct {
	Domain         string
	Secure         bool
	RedirectURI    string
	AuthStore      AuthStore
	SessionManager *scs.SessionManager
}

func GoogleConfig() *oauth2.Config {
	if googleOAuthConfig == nil {
		googleOAuthConfig = &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			Endpoint:     google.Endpoint,
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
			Scopes:       []string{"openid", "profile", "email"},
		}
	}

	return googleOAuthConfig
}

func generateState() (string, error) {
	nonceBytes := make([]byte, 64)
	_, err := io.ReadFull(rand.Reader, nonceBytes)

	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(nonceBytes), nil
}

func validateState(r *http.Request) error {
	state := r.URL.Query().Get("state")

	if state == "" {
		return errors.New("missing state")
	}

	cookie, err := r.Cookie(OAUTH_STATE_SESSION_KEY)

	if err != nil {
		return err
	}

	if cookie.Value != state {
		return ErrStateMismatch
	}

	return nil
}

func (ph *ProviderHandler) HandleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	verifier := oauth2.GenerateVerifier()

	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     OAUTH_STATE_SESSION_KEY,
		Value:    state,
		MaxAge:   60 * 5,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Domain:   ph.Domain,
		Secure:   ph.Secure,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     OAUTH_VERIFIER_SESSION_KEY,
		Value:    verifier,
		MaxAge:   60 * 5,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Domain:   ph.Domain,
		Secure:   ph.Secure,
	})

	url := GoogleConfig().AuthCodeURL(state, oauth2.AccessTypeOnline, oauth2.S256ChallengeOption(verifier))

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleToken is a custom struct to hold the oidc token response
// ExpiresIn remaining lifetime of the token in seconds
type GoogleToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
	IDToken      string    `json:"id_token"`
	ExpiresIn    *int      `json:"expires_in,omitempty"`
	Scope        string    `json:"scope"`
}

type googleUserInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	Locale        string `json:"locale"`
}

func newGoogleToken(tok *oauth2.Token) *GoogleToken {

	tokExpiresIn := tok.Extra("expires_in")

	if tokExpiresIn != nil {
		expiresIn := int(math.Round(tokExpiresIn.(float64)))

		return &GoogleToken{
			AccessToken:  tok.AccessToken,
			RefreshToken: tok.RefreshToken,
			Expiry:       tok.Expiry,
			IDToken:      tok.Extra("id_token").(string),
			ExpiresIn:    &expiresIn,
			Scope:        tok.Extra("scope").(string),
		}
	}

	return &GoogleToken{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
		IDToken:      tok.Extra("id_token").(string),
		ExpiresIn:    nil,
		Scope:        tok.Extra("scope").(string),
	}
}

// googleExchange exchanges the code for tokens and gets user info from Google
func googleExchange(ctx context.Context, code string, verifier string) (*GoogleToken, *googleUserInfo, error) {
	tok, err := GoogleConfig().Exchange(ctx, code, oauth2.VerifierOption(verifier))

	if err != nil {
		return nil, nil, err
	}

	googleToken := newGoogleToken(tok)

	client := GoogleConfig().Client(ctx, tok)

	userInfo, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	defer userInfo.Body.Close()

	if err != nil {
		return nil, nil, err
	}

	var userJson googleUserInfo
	json.NewDecoder(userInfo.Body).Decode(&userJson)

	return googleToken, &userJson, nil
}

func (ph *ProviderHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	reqLogger := internal.RequestLogger(r)

	code := r.URL.Query().Get("code")

	stateErr := validateState(r)

	if stateErr != nil {
		http.Error(w, stateErr.Error(), http.StatusBadRequest)
	}

	verifierCookie, err := r.Cookie(OAUTH_VERIFIER_SESSION_KEY)

	if err != nil {
		reqLogger.Error("Error getting verifier cookie", slog.String("err", err.Error()))
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	verifier := verifierCookie.Value

	if verifier == "" {
		reqLogger.Error("Verifier cookie is empty")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token, userJson, err := googleExchange(r.Context(), code, verifier)

	if err != nil {
		reqLogger.Error("Error exchanging code for token", slog.String("err", err.Error()))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Save tokens and user info to db if valid, save session, login
	// TODO: Check if user and account already exists, if so, update tokens and expiry, login

	existingUser, err := ph.AuthStore.GetUserAccount(r.Context(), userJson.Email, "google")
	accessTokenExpiryValid := token.Expiry != time.Time{}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			reqLogger.Info("User not found, creating new account")

			userId, createAccErr := ph.AuthStore.CreateAccount(r.Context(), CreateAccountData{
				Email:        userJson.Email,
				Provider:     "google",
				ProviderId:   userJson.Sub,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				AccessTokenExpiry: sql.NullTime{
					Time:  token.Expiry,
					Valid: accessTokenExpiryValid,
				},
			})

			if createAccErr != nil {
				reqLogger.Error("Error creating account", slog.String("err", createAccErr.Error()))

				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			createAccErr = Login(r, ph.SessionManager, SessionData{
				UserID: userId,
			})

			if createAccErr != nil {
				reqLogger.Error("Error logging in user", slog.String("err", createAccErr.Error()))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, ph.RedirectURI, http.StatusTemporaryRedirect)
			return
		} else {
			// Unexpected error

			reqLogger.Error("Failed OAuth callback", slog.String("err", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	err = ph.AuthStore.UpdateAccountTokens(r.Context(), existingUser.UserId, "google", userJson.Sub, UpdateAccountTokensData{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		AccessTokenExpiry: sql.NullTime{
			Time:  token.Expiry,
			Valid: accessTokenExpiryValid,
		},
	})

	if err != nil {
		reqLogger.Error("Error updating account tokens", slog.String("err", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = Login(r, ph.SessionManager, SessionData{
		UserID: existingUser.UserId,
	})

	if err != nil {
		reqLogger.Error("Error logging in user", slog.String("err", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, ph.RedirectURI, http.StatusTemporaryRedirect)
}
