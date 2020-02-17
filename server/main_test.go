package server

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFunctional(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "Done")
	})

	httpServer := httptest.NewUnstartedServer(handler)
	defer httpServer.Close()

	var logBuffer bytes.Buffer
	srv := NewServer(httpServer.Config, log.New(&logBuffer, "", 0))

	srv.Shutdown()
	srv.ListenAndServe(func() {})
}
