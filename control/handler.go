package control

import (
	"log"
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
		log.Fatalln("RPC register Responses error:", err)
	}
	if err := rpcServer.Register(NewRequests(queues.Requests)); err != nil {
		log.Fatalln("RPC register Requests error:", err)
	}

	return &control{
		queues:  queues,
		jsonrpc: NewJsonRPC(rpcServer),
	}
}

func (c *control) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
