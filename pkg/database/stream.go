package database

// Stream is an interface for streaming events from a collection.
type Stream interface {
	Next() <-chan Event    // Next returns a channel for receiving events.
	Done() <-chan struct{} // Done returns a channel indicating when streaming is done.
	Close() error          // Close closes the stream.
}
