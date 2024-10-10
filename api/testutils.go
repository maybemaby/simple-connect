package api

import (
	"net/http"
	"net/http/httptest"
)

func BootstrapTestHandler(path string, handler http.Handler) *httptest.Server {
	mux := http.NewServeMux()

	mux.Handle(path, handler)

	server := httptest.NewUnstartedServer(mux)

	server.EnableHTTP2 = true
	return server
}
