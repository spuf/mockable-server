package control

import (
	"net/http"
	"strings"
)

type Headers map[string]string

type Response struct {
	Status  int     `json:"status"`
	Headers Headers `json:"headers"`
	Body    string  `json:"body"`
}

type Request struct {
	Method  string  `json:"method"`
	Url     string  `json:"url"`
	Headers Headers `json:"headers"`
	Body    string  `json:"body"`
}

func (h *Headers) ToHttpHeaders() http.Header {
	res := make(http.Header, len(*h))
	for name, value := range *h {
		res.Set(name, value)
	}

	return res
}

func FromHttpHeaders(h http.Header) Headers {
	headers := make(Headers, len(h))
	for name, values := range h {
		headers[name] = strings.Join(values, "; ")
	}

	return headers
}
