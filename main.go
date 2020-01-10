package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"mockable-server/control"
	"mockable-server/server"
	"mockable-server/storage"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
)

type Queues struct {
	Responses storage.Store
	Requests  storage.Store
}

func main() {
	var mockAddr string
	var controlAddr string

	flag.StringVar(&mockAddr, "mock-addr", ":8000", "Mock server address")
	flag.StringVar(&controlAddr, "control-addr", ":8010", "Responses address")
	flag.Parse()

	done := make(chan struct{})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	queues := &Queues{
		Responses: storage.NewStore(),
		Requests:  storage.NewStore(),
	}

	controlServer := newControlServer(controlAddr, queues)
	mockServer := newMockServer(mockAddr, queues)

	go func() {
		<-quit
		controlServer.Shutdown()
		mockServer.Shutdown()
		close(done)
	}()

	go controlServer.ListenAndServe()
	go mockServer.ListenAndServe()

	<-done
}

func newMockServer(addr string, queues *Queues) *server.Server {
	logger := log.New(os.Stdout, "[mock]\t", log.LstdFlags)

	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res := queues.Responses.PopFirst()
		if res == nil {
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		}

		requestBody, _ := ioutil.ReadAll(r.Body)
		queues.Requests.PushLast(storage.Message{
			Headers: r.Header,
			Body:    string(requestBody),
			Request: &storage.Request{
				Method: r.Method,
				Url:    r.URL.RequestURI(),
			},
		})

		for name, values := range res.Headers {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		w.WriteHeader(res.Response.Status)
		_, err := io.WriteString(w, res.Body)
		if err != nil {
			logger.Fatal(err)
		}
	})

	return server.NewServer(addr, handler, logger)
}

func newControlServer(addr string, queues *Queues) *server.Server {
	logger := log.New(os.Stdout, "[control]\t", log.LstdFlags)

	rpcServer := rpc.NewServer()
	var err error
	err = rpcServer.Register(control.NewResponses(queues.Responses))
	if err != nil {
		logger.Fatal("RPC register Responses error:", err)
	}
	err = rpcServer.Register(control.NewRequests(queues.Requests))
	if err != nil {
		logger.Fatal("RPC register Requests error:", err)
	}

	jsonrpc := control.NewJsonRPC(rpcServer)
	handler := http.NewServeMux()
	handler.HandleFunc("/rpc/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			status := http.StatusMethodNotAllowed
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, http.StatusText(status), status)
			return
		}

		jsonrpc.Handle(w, r)
	})

	return server.NewServer(addr, handler, logger)
}
