package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"
)

func ListenAndServeWithGracefulShutdown(ctx context.Context, server *http.Server) {
	if server.IdleTimeout == 0 {
		server.IdleTimeout = time.Minute
	}
	if server.ErrorLog == nil {
		server.ErrorLog = log.New(os.Stdout, "server", log.LstdFlags)
	}

	c := make(chan error, 1)
	go func() {
		server.ErrorLog.Println("Starting server to listen on", server.Addr)
		c <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		server.ErrorLog.Println("Shutting down server")

		ctxWithTimeout, cancel := context.WithTimeout(ctx, server.IdleTimeout)
		defer cancel()

		if err := server.Shutdown(ctxWithTimeout); err != nil {
			server.ErrorLog.Fatalln("Could not gracefully shutdown server:", err)
		}

		if err := <-c; err != nil && !errors.Is(err, http.ErrServerClosed) {
			server.ErrorLog.Fatalln("Server failed to serve:", err)
		}
	case err := <-c:
		if err != nil {
			server.ErrorLog.Fatalln("Server failed to listen:", err)
		}
	}
}
