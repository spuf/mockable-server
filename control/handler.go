package control

import (
	"fmt"
	"net/http"
	"net/rpc"

	"github.com/spuf/mockable-server/storage"
)

type control struct {
	queues  *storage.Queues
	jsonrpc http.Handler
}

func NewHandler(queues *storage.Queues) http.Handler {
	rpcServer := rpc.NewServer()
	if err := rpcServer.Register(NewResponses(queues.Responses)); err != nil {
		panic(err)
	}
	if err := rpcServer.Register(NewRequests(queues.Requests)); err != nil {
		panic(err)
	}

	return &control{
		queues:  queues,
		jsonrpc: NewJsonRPC(rpcServer),
	}
}

func (c *control) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/healthz" {
		w.Header().Set("Content-Type", "text/plain")
		status := http.StatusOK
		w.WriteHeader(status)
		fmt.Fprintln(w, http.StatusText(status))
		return
	}

	if r.URL.Path != "/rpc/1" {
		status := http.StatusNotFound
		http.Error(w, http.StatusText(status), status)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
		return
	}

	c.jsonrpc.ServeHTTP(w, r)
}
