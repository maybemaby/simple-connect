package api

import (
	"context"
	v1 "simple-connect/gen/proto/api/v1"

	"connectrpc.com/connect"
)

type HealthService struct{}

func (hs *HealthService) Check(context.Context, *connect.Request[v1.Empty]) (*connect.Response[v1.CheckResponse], error) {
	return connect.NewResponse(&v1.CheckResponse{
		Message: "OK",
	}), nil
}
