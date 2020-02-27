package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerHandler(t *testing.T) {
	var b bytes.Buffer
	logger := log.New(&b, "Header:", 0)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	server := httptest.NewServer(NewLoggerHandler(logger, handler))

	defer server.Close()

	if _, err := http.Get(server.URL); err != nil {
		t.Fatal(err)
	}

	// @todo: add assertions
}
