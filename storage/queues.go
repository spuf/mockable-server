package storage

type Queues struct {
	Responses Store
	Requests  Store
}

func NewQueues() *Queues {
	return &Queues{
		Responses: NewStore(),
		Requests:  NewStore(),
	}
}
