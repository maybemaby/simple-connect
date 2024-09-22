package auth

import (
	"context"
	"net/http"
	"simple-connect/api/httputils"
	v1 "simple-connect/gen/proto/api/v1"

	"connectrpc.com/connect"
	"github.com/alexedwards/scs/v2"
)

type AuthHandler struct {
	store          AuthStore
	sessionManager *scs.SessionManager
}

type LoginResponse struct {
	Id string `json:"id"`
}

func NewAuthHandler(store AuthStore, sessionManager *scs.SessionManager) *AuthHandler {
	return &AuthHandler{store: store, sessionManager: sessionManager}
}

func (as *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	err := Login(r, as.sessionManager, "1")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputils.WriteJSON(w, r, LoginResponse{Id: "1"})
}

type ProtectedAuthHandler struct {
	store          AuthStore
	sessionManager *scs.SessionManager
}

func NewProtectedAuthHandler(store AuthStore, sessionManager *scs.SessionManager) *ProtectedAuthHandler {
	return &ProtectedAuthHandler{store: store, sessionManager: sessionManager}
}

func (as *ProtectedAuthHandler) Me(ctx context.Context, req *connect.Request[v1.MeRequest]) (*connect.Response[v1.BaseUser], error) {

	return connect.NewResponse(&v1.BaseUser{
		Id: "1",
	}), nil
}

func (as *ProtectedAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {

	Logout(r, as.sessionManager)

	w.WriteHeader(http.StatusOK)
}
