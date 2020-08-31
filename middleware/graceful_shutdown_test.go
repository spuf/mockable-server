package middleware

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"
)

func TestListenAndServeWithGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	srv := &http.Server{Addr: "127.0.0.1:0"}

	isShutdown := false
	var err error
	wg := new(sync.WaitGroup)

	wg.Add(1)
	srv.RegisterOnShutdown(func() {
		defer wg.Done()
		isShutdown = true
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = ListenAndServeWithGracefulShutdown(ctx, srv, func(_ net.Addr) {
			defer cancel()
		})
	}()

	wg.Wait()
	if err != nil {
		t.Logf("ListenAndServeWithGracefulShutdown: %v", err)
	}
	if !isShutdown {
		t.Errorf("Server was shutted down")
	}
}
