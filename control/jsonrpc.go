package control

import (
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type wrapper struct {
	body io.Reader
	w    io.Writer
	done chan struct{}
}

func (r *wrapper) Read(p []byte) (n int, err error) {
	return r.body.Read(p)
}

func (r *wrapper) Write(p []byte) (n int, err error) {
	return r.w.Write(p)
}

func (r *wrapper) Close() error {
	close(r.done)
	return nil
}

type jsonRPC struct {
	server *rpc.Server
}

func NewJsonRPC(server *rpc.Server) *jsonRPC {
	return &jsonRPC{server: server}
}

func (j *jsonRPC) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	wrap := &wrapper{
		body: r.Body,
		w:    w,
		done: make(chan struct{}),
	}
	go j.server.ServeCodec(jsonrpc.NewServerCodec(wrap))
	<-wrap.done
}
