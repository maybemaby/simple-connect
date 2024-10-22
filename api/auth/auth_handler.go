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
	"github.com/maybemaby/smolauth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthHandler struct {
	store       AuthStore
	authManager *smolauth.AuthManager
}

type LoginResponse struct {
	Id int `json:"id"`
}

type SignupRequest struct {
	Email     string `json:"email"`
	Password1 string `json:"password1"`
	Password2 string `json:"password2"`
}

func NewAuthHandler(store AuthStore, sessionManager *smolauth.AuthManager) *AuthHandler {
	return &AuthHandler{store: store, authManager: sessionManager}
}

func (as *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	logger := internal.RequestLogger(r)
	loginReq := &v1.LoginRequest{}

	err := httputils.ReadJSON(r, loginReq)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := as.authManager.CheckPassword(loginReq.Email, loginReq.Password)

	if err != nil {
		logger.Error("error checking password", slog.String("Error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Invalid email or password"))
		return
	}

	err = as.authManager.Login(r, smolauth.SessionData{UserId: id})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputils.WriteJSON(w, r, LoginResponse{Id: id})
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

	id, err := as.authManager.PasswordSignup(signupReq.Email, signupReq.Password1)

	if err != nil {
		logger.Error("error creating user", slog.String("Error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = as.authManager.Login(r, smolauth.SessionData{UserId: id})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputils.WriteJSON(w, r, LoginResponse{Id: id})
}

type ProtectedAuthHandler struct {
	store       AuthStore
	authManager *smolauth.AuthManager
}

func NewProtectedAuthHandler(store AuthStore, authManager *smolauth.AuthManager) *ProtectedAuthHandler {
	return &ProtectedAuthHandler{store: store, authManager: authManager}
}

func (as *ProtectedAuthHandler) Me(ctx context.Context, req *connect.Request[v1.MeRequest]) (*connect.Response[v1.ReadUser], error) {

	user, err := as.authManager.GetUserCtx(ctx)

	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
	}

	return connect.NewResponse(&v1.ReadUser{
		Id:        int64(user.Id),
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}), nil
}

func (as *ProtectedAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {

	as.authManager.Logout(r)

	w.WriteHeader(http.StatusOK)
}
