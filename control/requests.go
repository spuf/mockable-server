package control

import (
	"mockable-server/storage"
)

type Requests struct {
	store storage.Store
}

func NewRequests(store storage.Store) *Requests {
	return &Requests{store: store}
}

func (r *Requests) List(_ struct{}, reply *[]Request) error {
	list := r.store.List()
	res := make([]Request, len(list))
	for i, msg := range list {
		res[i] = Request{
			Method:  msg.Request.Method,
			Url:     msg.Request.Url,
			Headers: FromHttpHeaders(msg.Headers),
			Body:    msg.Body,
		}
	}

	*reply = res
	return nil
}
