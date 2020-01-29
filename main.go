package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"mockable-server/control"
	"mockable-server/server"
	"mockable-server/storage"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"sync"
)

type Queues struct {
	Responses storage.Store
	Requests  storage.Store
}

func main() {
	var mockAddr string
	var controlAddr string

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	queues := &Queues{
		Responses: storage.NewStore(),
		Requests:  storage.NewStore(),
	}

	servers := []*server.Server{
		newControlServer(controlAddr, queues),
		newMockServer(mockAddr, queues),
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

func newMockServer(addr string, queues *Queues) *server.Server {
	logger := log.New(os.Stdout, "[mock]\t", log.Lshortfile)

	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var body bytes.Buffer
		if _, err := body.ReadFrom(r.Body); err != nil {
			logger.Fatalln(err)
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
			logger.Fatalln(err)
		}
	})

	return server.NewServer(addr, handler, logger)
}

func newControlServer(addr string, queues *Queues) *server.Server {
	logger := log.New(os.Stdout, "[control]\t", log.Lshortfile)

	rpcServer := rpc.NewServer()
	if err := rpcServer.Register(control.NewResponses(queues.Responses)); err != nil {
		logger.Fatalln("RPC register Responses error:", err)
	}
	if err := rpcServer.Register(control.NewRequests(queues.Requests)); err != nil {
		logger.Fatalln("RPC register Requests error:", err)
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

	return server.NewServer(addr, handler, logger)
}
