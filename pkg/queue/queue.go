package queue

import "context"

type Job interface {
	Type() string
}

type Handler interface {
	Handle(context.Context, Job) error
}

type Queue interface {
	Enqueue(context.Context, Job) error
	Consume(context.Context) (<-chan Job, error)
	Close() error
}
