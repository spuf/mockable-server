package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	httpServer *http.Server
	logger     *log.Logger
}

type logEntry struct {
	Method  string
	URI     string
	Headers map[string]string
	Body    string
}

func NewServer(addr string, handler http.Handler, logger *log.Logger) *Server {
	logger.Println("Initialize server")

	loggableHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := make(map[string]string, len(r.Header))
		for name, values := range r.Header {
			headers[name] = strings.Join(values, "; ")
		}

		var body bytes.Buffer
		if _, err := body.ReadFrom(r.Body); err != nil {
			logger.Fatalln(err)
		}
		if err := r.Body.Close(); err != nil {
			logger.Fatalln(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(body.Bytes()))

		line, err := json.Marshal(logEntry{
			Method:  r.Method,
			URI:     r.URL.RequestURI(),
			Headers: headers,
			Body:    body.String(),
		})
		if err != nil {
			logger.Fatalln(err)
		}
		logger.Println(string(line))

		handler.ServeHTTP(w, r)
	})

	httpServer := &http.Server{
		Addr:        addr,
		Handler:     loggableHandler,
		ErrorLog:    logger,
		IdleTimeout: time.Minute,
	}
	server := &Server{
		httpServer: httpServer,
		logger:     logger,
	}

	return server
}

func (s *Server) ListenAndServe(onClose func()) {
	s.logger.Println("Starting server to listen on", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			onClose()
		} else {
			s.logger.Fatalln("Could not listen:", err)
		}
	}
}

func (s *Server) Shutdown() {
	s.logger.Println("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), s.httpServer.IdleTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Fatalln("Could not gracefully shutdown the server:", err)
	}
}
