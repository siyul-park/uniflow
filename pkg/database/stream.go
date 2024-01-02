package database

// Stream is an interface for streaming events from a collection.
type Stream interface {
	Next() <-chan Event
	Done() <-chan struct{}
	Close() error
}
