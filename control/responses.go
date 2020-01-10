package control

import (
	"errors"
	"mockable-server/storage"
)

type Responses struct {
	store storage.Store
}

func NewResponses(store storage.Store) *Responses {
	return &Responses{store: store}
}

func (r *Responses) List(_ struct{}, reply *[]Response) error {
	list := r.store.List()
	res := make([]Response, len(list))
	for i, msg := range list {
		res[i] = Response{
			Status:  msg.Response.Status,
			Headers: FromHttpHeaders(msg.Headers),
			Body:    msg.Body,
		}
	}

	*reply = res
	return nil
}

func (r *Responses) Push(arg Response, reply *bool) error {
	if arg.Status < 100 || arg.Status >= 600 {
		return errors.New("status must be in [100; 599]")
	}

	r.store.PushLast(storage.Message{
		Headers:  arg.Headers.ToHttpHeaders(),
		Body:     arg.Body,
		Response: &storage.Response{Status: arg.Status},
	})

	*reply = true
	return nil
}

func (r *Responses) Clear(_ struct{}, reply *bool) error {
	r.store.Clear()

	*reply = true
	return nil
}
