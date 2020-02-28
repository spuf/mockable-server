package control

import (
	"errors"
	"fmt"
	"github.com/spuf/mockable-server/storage"
	"net/http"
	"strings"
)

var ErrValidation = errors.New("validation")

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

type Headers map[string]string

func (h *Headers) ToHttpHeaders() http.Header {
	res := make(http.Header, len(*h))
	for name, value := range *h {
		res.Set(name, value)
	}

	return res
}

func fromHttpHeaders(h http.Header) Headers {
	headers := make(Headers, len(h))
	for name, values := range h {
		headers[name] = strings.Join(values, "; ")
	}

	return headers
}

func requestFromMessage(msg storage.Message) (*Request, error) {
	if !msg.IsRequest() {
		return nil, fmt.Errorf("%#v is not request", msg)
	}

	request := Request{
		Method:  msg.Request.Method,
		Url:     msg.Request.Url,
		Headers: fromHttpHeaders(msg.Headers),
		Body:    msg.Body,
	}
	return &request, nil
}
