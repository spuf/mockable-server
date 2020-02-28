package mock

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/spuf/mockable-server/storage"
)

type mock struct {
	queues *storage.Queues
}

func NewHandler(queues *storage.Queues) http.Handler {
	return &mock{queues: queues}
}

func (m *mock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body bytes.Buffer
	if _, err := body.ReadFrom(r.Body); err != nil {
		log.Fatalln(err)
	}

	message := storage.Message{
		Headers: r.Header,
		Body:    body.String(),
		Request: &storage.Request{
			Method: r.Method,
			Url:    r.URL.RequestURI(),
		},
	}
	if err := m.queues.Requests.PushLast(message); err != nil {
		log.Fatalln(err)
	}

	res := m.queues.Responses.PopFirst()
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
}
