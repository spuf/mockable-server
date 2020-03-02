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

func (m Message) IsRequest() bool {
	return m.Request != nil && m.Response == nil
}

func (m Message) IsResponse() bool {
	return m.Response != nil && m.Request == nil
}

type store struct {
	items     []*Message
	validator func(Message) error
}

type Store interface {
	PushLast(message Message) error
	PopFirst() *Message
	List() []Message
	Clear()
}

func (s *store) PushLast(message Message) error {
	if s.validator != nil {
		if err := s.validator(message); err != nil {
			return err
		}
	}

	s.items = append(s.items, &message)

	return nil
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
	s.items = nil
}

func NewStore(validator func(Message) error) Store {
	return &store{validator: validator}
}
