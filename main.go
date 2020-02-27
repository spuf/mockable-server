package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/spuf/mockable-server/control"
	"github.com/spuf/mockable-server/server"
	"github.com/spuf/mockable-server/storage"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"sync"
)

var (
	Application = "mockable-server"
	Version     = ""
)

type Queues struct {
	Responses storage.Store
	Requests  storage.Store
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s (%s):\n", Application, Version)
		flag.PrintDefaults()
	}

	mockAddr := flag.String("mock-addr", ":8010", "Mock server address")
	controlAddr := flag.String("control-addr", ":8020", "Control server address")

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	queues := &Queues{
		Responses: storage.NewStore(),
		Requests:  storage.NewStore(),
	}

	servers := []*server.Server{
		server.NewServer(&http.Server{Addr: *controlAddr, Handler: newControlHandler(queues)}, log.New(os.Stdout, "[mock]\t", 0)),
		server.NewServer(&http.Server{Addr: *mockAddr, Handler: newMockHandle(queues)}, log.New(os.Stdout, "[control]\t", 0)),
	}

	go func() {
		<-quit
		for _, srv := range servers {
			go srv.Shutdown()
		}
	}()

	wg := new(sync.WaitGroup)
	for _, srv := range servers {
		wg.Add(1)
		go srv.ListenAndServe(func() {
			wg.Done()
		})
	}

	wg.Wait()
}

func newMockHandle(queues *Queues) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body bytes.Buffer
		if _, err := body.ReadFrom(r.Body); err != nil {
			log.Fatalln(err)
		}

		queues.Requests.PushLast(storage.Message{
			Headers: r.Header,
			Body:    body.String(),
			Request: &storage.Request{
				Method: r.Method,
				Url:    r.URL.RequestURI(),
			},
		})

		res := queues.Responses.PopFirst()
		if res == nil {
			status := http.StatusNotImplemented
			http.Error(w, http.StatusText(status), status)
			return
		}

		for name, values := range res.Headers {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		w.WriteHeader(res.Response.Status)
		if _, err := io.WriteString(w, res.Body); err != nil {
			log.Fatalln(err)
		}
	})

	return handler
}

func newControlHandler(queues *Queues) http.Handler {
	rpcServer := rpc.NewServer()
	if err := rpcServer.Register(control.NewResponses(queues.Responses)); err != nil {
		log.Fatalln("RPC register Responses error:", err)
	}
	if err := rpcServer.Register(control.NewRequests(queues.Requests)); err != nil {
		log.Fatalln("RPC register Requests error:", err)
	}

	jsonrpc := control.NewJsonRPC(rpcServer)
	handler := http.NewServeMux()
	handler.HandleFunc("/rpc/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			status := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(status), status)
			return
		}

		jsonrpc.ServeHTTP(w, r)
	})

	return handler
}
