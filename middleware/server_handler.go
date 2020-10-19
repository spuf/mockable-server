package middleware

import (
	"net/http"
)

type serverHandler struct {
	server string
	next   http.Handler
}

func NewServerHandler(server string, next http.Handler) http.Handler {
	return &serverHandler{
		server: server,
		next:   next,
	}
}

func (m *serverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", m.server)
	m.next.ServeHTTP(w, r)
}
