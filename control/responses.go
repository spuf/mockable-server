package control

import (
	"fmt"
	"github.com/spuf/mockable-server/storage"
)

type Responses struct {
	store storage.Store
}

func NewResponses(store storage.Store) *Responses {
	return &Responses{store: store}
}

func (r *Responses) List(_ struct{}, reply *[]Response) error {
	list := r.store.List()
	for _, msg := range list {
		response := Response{
			Status:  msg.Response.Status,
			Headers: fromHttpHeaders(msg.Headers),
			Body:    msg.Body,
		}
		*reply = append(*reply, response)
	}

	return nil
}

func (r *Responses) Push(arg Response, reply *bool) error {
	if arg.Status < 100 || arg.Status >= 600 {
		return fmt.Errorf("%w: status %v must be in [100; 600)", ErrValidation, arg.Status)
	}

	msg := storage.Message{
		Headers:  arg.Headers.ToHttpHeaders(),
		Body:     arg.Body,
		Response: &storage.Response{Status: arg.Status},
	}
	if err := r.store.PushLast(msg); err != nil {
		return err
	}

	*reply = true

	return nil
}

func (r *Responses) Clear(_ struct{}, reply *bool) error {
	r.store.Clear()
	*reply = true

	return nil
}
