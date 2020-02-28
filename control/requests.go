package control

import (
	"github.com/spuf/mockable-server/storage"
)

type Requests struct {
	store storage.Store
}

func NewRequests(store storage.Store) *Requests {
	return &Requests{store: store}
}

func (r *Requests) List(_ struct{}, reply *[]Request) error {
	list := r.store.List()
	for _, msg := range list {
		if request, err := requestFromMessage(msg); err != nil {
			return err
		} else {
			*reply = append(*reply, *request)
		}
	}

	return nil
}

func (r *Requests) Pop(_ struct{}, reply *interface{}) error {
	if msg := r.store.PopFirst(); msg != nil {
		if request, err := requestFromMessage(*msg); err != nil {
			return err
		} else {
			*reply = *request
		}
	}

	return nil
}

func (r *Requests) Clear(_ struct{}, reply *bool) error {
	r.store.Clear()
	*reply = true

	return nil
}
