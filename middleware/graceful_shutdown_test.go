package middleware

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"strings"
	"testing"
)

func TestListenAndServeWithGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buffer bytes.Buffer
	srv := &http.Server{
		ErrorLog: log.New(&buffer, "", 0),
	}

	ListenAndServeWithGracefulShutdown(ctx, srv)

	assertLogsContains := func(t *testing.T, logs, substr string) {
		t.Helper()
		if !strings.Contains(logs, substr) {
			t.Errorf("Log does not contain:\n got: %#v\nwant substring: %#v", logs, substr)
		}
	}

	assertLogsContains(t, buffer.String(), "Starting server to listen on")
	assertLogsContains(t, buffer.String(), "Shutting down server")
}
