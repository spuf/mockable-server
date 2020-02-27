package middleware

import (
	"context"
	"net/http"
	"testing"
)

func TestListenAndServeWithGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ListenAndServeWithGracefulShutdown(ctx, &http.Server{})

	// @todo: add assertions
}
