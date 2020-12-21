package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type loggerHandler struct {
	logger *log.Logger
	next   http.Handler
}

type logEntry struct {
	Method  string
	URI     string
	Headers map[string]string
	Body    string
}

func NewLoggerHandler(logger *log.Logger, next http.Handler) http.Handler {
	return &loggerHandler{
		logger: logger,
		next:   next,
	}
}

func (m *loggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	headers := make(map[string]string, len(r.Header))
	for name, values := range r.Header {
		headers[name] = strings.Join(values, "; ")
	}

	body, err := m.drainBody(r)
	if err != nil {
		panic(err)
	}

	entry := logEntry{
		Method:  r.Method,
		URI:     r.URL.RequestURI(),
		Headers: headers,
		Body:    body.String(),
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(entry); err != nil {
		panic(err)
	}

	m.logger.Printf("%s", buf.String())

	m.next.ServeHTTP(w, r)
}

func (m *loggerHandler) drainBody(r *http.Request) (*bytes.Buffer, error) {
	var body bytes.Buffer
	if _, err := body.ReadFrom(r.Body); err != nil {
		return nil, err
	}
	if err := r.Body.Close(); err != nil {
		return nil, err
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(body.Bytes()))

	return &body, nil
}
