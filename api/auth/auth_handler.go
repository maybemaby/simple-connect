package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"simple-connect/api/httputils"
	"simple-connect/api/internal"
	v1 "simple-connect/gen/proto/api/v1"

	"connectrpc.com/connect"
	"github.com/alexedwards/scs/v2"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthHandler struct {
	store          AuthStore
	sessionManager *scs.SessionManager
}

type LoginResponse struct {
	Id string `json:"id"`
}

type SignupRequest struct {
	Email     string `json:"email"`
	Password1 string `json:"password1"`
	Password2 string `json:"password2"`
}

func NewAuthHandler(store AuthStore, sessionManager *scs.SessionManager) *AuthHandler {
	return &AuthHandler{store: store, sessionManager: sessionManager}
}

func (as *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	logger := internal.RequestLogger(r)
	loginReq := &v1.LoginRequest{}

	err := httputils.ReadJSON(r, loginReq)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := as.store.GetUserByEmail(r.Context(), loginReq.Email)

	if err != nil {
		logger.Error("error getting user by email", slog.String("Error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(loginReq.Password))

	if err != nil {
		logger.Debug("passwords do not match", slog.String("pw", user.PasswordHash.String))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = Login(r, as.sessionManager, SessionData{UserID: user.ID})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputils.WriteJSON(w, r, LoginResponse{Id: user.ID})
}

func (as *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	logger := internal.RequestLogger(r)
	signupReq := &SignupRequest{}

	err := httputils.ReadJSON(r, signupReq)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if signupReq.Password1 != signupReq.Password2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := as.store.CreateUser(r.Context(), signupReq.Email, signupReq.Password1)

	if err != nil {
		logger.Error("error creating user", slog.String("Error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = Login(r, as.sessionManager, SessionData{UserID: id})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputils.WriteJSON(w, r, LoginResponse{Id: id})
}

type ProtectedAuthHandler struct {
	store          AuthStore
	sessionManager *scs.SessionManager
}

func NewProtectedAuthHandler(store AuthStore, sessionManager *scs.SessionManager) *ProtectedAuthHandler {
	return &ProtectedAuthHandler{store: store, sessionManager: sessionManager}
}

func (as *ProtectedAuthHandler) Me(ctx context.Context, req *connect.Request[v1.MeRequest]) (*connect.Response[v1.ReadUser], error) {

	userId := as.sessionManager.GetString(ctx, SessionUserKey)

	if userId == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
	}

	user, err := as.store.GetUserByID(ctx, userId)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.ReadUser{
		Id:        user.ID,
		Email:     *user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}), nil
}

func (as *ProtectedAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {

	Logout(r, as.sessionManager)

	w.WriteHeader(http.StatusOK)
}
