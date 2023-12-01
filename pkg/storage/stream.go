package storage

import (
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type (
	// Stream is a stream to track scheme.Spec changes.
	Stream struct {
		stream  database.Stream 
		channel chan Event     
		done    chan struct{}
	}
)

// NewStream returns a new Stream.
func NewStream(stream database.Stream) *Stream {
	s := &Stream{
		stream:  stream,
		channel: make(chan Event),
		done:    make(chan struct{}),
	}

	go func() {
		defer func() { close(s.channel) }()

		for {
			select {
			case <-s.done:
				return
			case <-s.stream.Done():
				_ = s.Close()
				return
			case e := <-s.stream.Next():
				var id ulid.ULID
				if err := primitive.Unmarshal(e.DocumentID, &id); err != nil {
					continue
				}
				var op eventOP
				if e.OP == database.EventInsert {
					op = EventInsert
				} else if e.OP == database.EventUpdate {
					op = EventUpdate
				} else if e.OP == database.EventDelete {
					op = EventDelete
				}

				select {
				case <-s.done:
					return
				case s.channel <- Event{OP: op, NodeID: id}:
				}
			}
		}
	}()

	return s
}

// Next returns a channel that receives Event.
func (s *Stream) Next() <-chan Event {
	return s.channel
}

// Done returns a channel that is closed when the Stream is closed.
func (s *Stream) Done() <-chan struct{} {
	return s.done
}

// Close closes the Stream.
func (s *Stream) Close() error {
	select {
	case <-s.done:
		return nil
	default:
	}

	close(s.done)

	return s.stream.Close()
}
