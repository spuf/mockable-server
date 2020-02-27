package storage

import (
	"net/http"
)

type Request struct {
	Method string
	Url    string
}
type Response struct {
	Status int
}
type Message struct {
	Headers http.Header
	Body    string

	Request  *Request
	Response *Response
}

type store struct {
	items []*Message
}

type Store interface {
	PushLast(message Message)
	PopFirst() *Message
	List() []Message
	Clear()
}

func (s *store) PushLast(message Message) {
	s.items = append(s.items, &message)
}

func (s *store) PopFirst() *Message {
	if len(s.items) <= 0 {
		return nil
	}

	item := s.items[0]
	s.items = s.items[1:]

	return item
}
func (s *store) List() []Message {
	res := make([]Message, len(s.items))
	for i, mes := range s.items {
		res[i] = *mes
	}

	return res
}

func (s *store) Clear() {
	s.items = []*Message{}
}

func NewStore() Store {
	return &store{}
}
