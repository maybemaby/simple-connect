package api

import (
	"context"
	apiv1 "simple-connect/gen/proto/api/v1"
	"simple-connect/gen/proto/api/v1/apiv1connect"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler(t *testing.T) {

	t.Parallel()

	server := BootstrapTestHandler(apiv1connect.NewHealthServiceHandler(&HealthService{}))

	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	t.Run("HealthService.Check", func(t *testing.T) {

		client := apiv1connect.NewHealthServiceClient(server.Client(), server.URL)

		resp, err := client.Check(context.Background(), &connect.Request[apiv1.Empty]{})

		assert.NoError(t, err)

		assert.Equal(t, "OK", resp.Msg.Message)
	})
}
