package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/spuf/mockable-server/control"
	"github.com/spuf/mockable-server/middleware"
	"github.com/spuf/mockable-server/mock"
	"github.com/spuf/mockable-server/storage"
)

var (
	Application = "mockable-server"
	Version     string
	mockAddr    string
	controlAddr string
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s (%s):\n", Application, Version)
		flag.PrintDefaults()
	}

	flag.StringVar(&mockAddr, "mock-addr", ":8010", "Mock server address")
	flag.StringVar(&controlAddr, "control-addr", ":8020", "Control server address")

	flag.VisitAll(func(f *flag.Flag) {
		envName := strings.ReplaceAll(strings.ToUpper(f.Name), "-", "_")
		if envVal, ok := os.LookupEnv(envName); ok {
			if err := flag.Set(f.Name, envVal); err != nil {
				panic(err)
			}
		}

		f.Usage = fmt.Sprintf("%s [%s]", f.Usage, envName)
	})
	flag.Parse()

	logFlags := log.LstdFlags | log.Lmsgprefix
	if Version == "" {
		logFlags = logFlags | log.Lshortfile
	}
	controlLogger := log.New(os.Stdout, "[control] ", logFlags)
	mockLogger := log.New(os.Stdout, "[mock] ", logFlags)

	queues := storage.NewQueues()
	servers := [...]*http.Server{
		{
			Addr: controlAddr,
			Handler: middleware.NewServerHandler(fmt.Sprintf("%s %s (control)", Application, Version),
				middleware.NewLoggerHandler(controlLogger,
					control.NewHandler(queues))),
			ErrorLog: controlLogger,
		},
		{
			Addr: mockAddr,
			Handler: middleware.NewServerHandler(fmt.Sprintf("%s %s", Application, Version),
				middleware.NewLoggerHandler(mockLogger,
					mock.NewHandler(queues))),
			ErrorLog: mockLogger,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErrors := make(chan error, len(servers))
	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt)
	go func() {
		defer cancel()
		<-quitSignal
	}()

	wg := new(sync.WaitGroup)
	for _, srv := range servers {
		wg.Add(1)
		go func(srv *http.Server) {
			defer wg.Done()

			srv.RegisterOnShutdown(func() {
				if srv.ErrorLog != nil {
					srv.ErrorLog.Printf("Shutting down server")
				}
			})

			onListen := func(addr net.Addr) {
				if srv.ErrorLog != nil {
					srv.ErrorLog.Printf("Server is listening on %s", addr.String())
				}
			}

			if err := middleware.ListenAndServeWithGracefulShutdown(ctx, srv, onListen); err != nil {
				defer cancel()
				serverErrors <- err
				if srv.ErrorLog != nil {
					srv.ErrorLog.Printf("failed: %v", err)
				}
			}
		}(srv)
	}
	wg.Wait()

	select {
	case err := <-serverErrors:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	default:
		close(serverErrors)
	}
}
