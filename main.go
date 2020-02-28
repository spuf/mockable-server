package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/spuf/mockable-server/control"
	"github.com/spuf/mockable-server/middleware"
	"github.com/spuf/mockable-server/mock"
	"github.com/spuf/mockable-server/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
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
				log.Fatalln(err)
			}
		}

		f.Usage = fmt.Sprintf("%s [%s]", f.Usage, envName)
	})
	flag.Parse()

	controlLogger := log.New(os.Stdout, "[control]\t", 0)
	mockLogger := log.New(os.Stdout, "[mock]\t", 0)

	queues := storage.NewQueues()
	servers := [...]*http.Server{
		&http.Server{
			Addr:     controlAddr,
			Handler:  middleware.NewLoggerHandler(controlLogger, control.NewHandler(queues)),
			ErrorLog: controlLogger,
		},
		&http.Server{
			Addr:     mockAddr,
			Handler:  middleware.NewLoggerHandler(mockLogger, mock.NewHandler(queues)),
			ErrorLog: mockLogger,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		defer cancel()
		<-quit
	}()

	wg := new(sync.WaitGroup)
	for _, srv := range servers {
		wg.Add(1)
		go func(srv *http.Server) {
			defer wg.Done()
			middleware.ListenAndServeWithGracefulShutdown(ctx, srv)
		}(srv)
	}
	wg.Wait()
}
