package control

import (
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type readWriteCloser struct {
	readCloser io.ReadCloser
	writer     io.Writer
}

func (r *readWriteCloser) Read(p []byte) (n int, err error) {
	return r.readCloser.Read(p)
}

func (r *readWriteCloser) Write(p []byte) (n int, err error) {
	return r.writer.Write(p)
}

func (r *readWriteCloser) Close() error {
	return r.readCloser.Close()
}

type jsonRPC struct {
	server *rpc.Server
}

func NewJsonRPC(server *rpc.Server) *jsonRPC {
	return &jsonRPC{server: server}
}

func (j *jsonRPC) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	codec := jsonrpc.NewServerCodec(&readWriteCloser{r.Body, w})
	defer codec.Close()

	_ = j.server.ServeRequest(codec)
}
