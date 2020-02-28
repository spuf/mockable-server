package storage

import (
	"testing"
)

func TestQueues(t *testing.T) {
	queues := NewQueues()

	req := Message{
		Request: &Request{},
	}
	res := Message{
		Response: &Response{},
	}

	if err := queues.Responses.PushLast(req); err == nil {
		t.Errorf("PushLast must return error")
	}
	if err := queues.Responses.PushLast(res); err != nil {
		t.Errorf("PushLast must not return error: %v", err)
	}

	if err := queues.Requests.PushLast(req); err != nil {
		t.Errorf("PushLast must not return error: %v", err)
	}
	if err := queues.Requests.PushLast(res); err == nil {
		t.Errorf("PushLast must return error")
	}
}
