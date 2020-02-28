package storage

import "fmt"

type Queues struct {
	Responses Store
	Requests  Store
}

func NewQueues() *Queues {
	return &Queues{
		Responses: NewStore(responseValidator),
		Requests:  NewStore(requestValidator),
	}
}

func responseValidator(message Message) error {
	if !message.IsResponse() {
		return fmt.Errorf("%#v is not Response", message)
	}

	return nil
}

func requestValidator(message Message) error {
	if !message.IsRequest() {
		return fmt.Errorf("%#v is not Request", message)
	}

	return nil
}
