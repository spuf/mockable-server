package control

import (
	"encoding/base64"
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
			Delay:   DelayDuration{msg.Delay},
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
		return fmt.Errorf("%w: status %d must be in [100; 600)", ErrValidation, arg.Status)
	}

	body := arg.Body
	if arg.IsBodyBase64 {
		decodedBody, err := base64.StdEncoding.DecodeString(arg.Body)
		if err != nil {
			return fmt.Errorf("failed to decode body from base64: %w", err)
		}
		body = string(decodedBody)
	}

	msg := storage.Message{
		Delay:    arg.Delay.Duration,
		Headers:  arg.Headers.ToHttpHeaders(),
		Body:     body,
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
