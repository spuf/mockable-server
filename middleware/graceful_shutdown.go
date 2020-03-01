package middleware

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

func ListenAndServeWithGracefulShutdown(ctx context.Context, server *http.Server, onListen func(net.Addr)) error {
	if server.Addr == "" {
		return fmt.Errorf("server addr must be defined")
	}
	if server.IdleTimeout == 0 {
		server.IdleTimeout = time.Minute
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	if onListen != nil {
		onListen(ln.Addr())
	}

	c := make(chan error, 1)
	go func() {
		c <- server.Serve(ln)
	}()

	select {
	case <-ctx.Done():
		ctxWithTimeout, cancel := context.WithTimeout(ctx, server.IdleTimeout)
		defer cancel()

		if err := server.Shutdown(ctxWithTimeout); err != nil {
			return fmt.Errorf("could not gracefully shutdown server: %w", err)
		}

		if err := <-c; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server failed to serve: %w", err)
		}

	case err := <-c:
		if err != nil {
			return fmt.Errorf("server failed to listen: %w", err)
		}
	}

	return nil
}
