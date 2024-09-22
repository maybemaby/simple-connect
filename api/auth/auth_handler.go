package auth

import (
	"context"
	v1 "simple-connect/gen/proto/api/v1"

	"connectrpc.com/connect"
)

type AuthHandler struct {
	store *AuthStore
}

func NewAuthHandler(store *AuthStore) *AuthHandler {
	return &AuthHandler{store: store}
}

func (as *AuthHandler) Login(ctx context.Context, req *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {

	res := connect.NewResponse(&v1.LoginResponse{})

	return res, nil
}

type ProtectedAuthHandler struct{}

func (as *ProtectedAuthHandler) Me(ctx context.Context, req *connect.Request[v1.MeRequest]) (*connect.Response[v1.BaseUser], error) {

	return connect.NewResponse(&v1.BaseUser{
		Id: "1",
	}), nil
}
