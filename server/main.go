package server

import (
	"context"
	"errors"
	"log"
	"net/http"
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

func NewServer(httpServer *http.Server, logger *log.Logger) *Server {
	logger.Println("Initialize server")

	httpServer.Handler = NewLoggerMiddleware(logger, httpServer.Handler)
	if httpServer.IdleTimeout == 0 {
		httpServer.IdleTimeout = time.Minute
	}
	if httpServer.ErrorLog == nil {
		httpServer.ErrorLog = logger
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
