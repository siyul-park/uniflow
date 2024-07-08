package store

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/types"
)

// Stream represents a stream for tracking spec.Spec changes.
type Stream struct {
	stream  database.Stream
	channel chan Event
}

// newStream creates a new Stream based on the provided database.Stream.
func newStream(stream database.Stream) *Stream {
	s := &Stream{
		stream:  stream,
		channel: make(chan Event),
	}

	go func() {
		defer close(s.channel)

		for {
			select {
			case <-s.stream.Done():
				return
			case e := <-s.stream.Next():
				var id uuid.UUID
				if err := types.Decoder.Decode(e.DocumentID, &id); err != nil {
					continue
				}

				var op eventOP
				switch e.OP {
				case database.EventInsert:
					op = EventInsert
				case database.EventUpdate:
					op = EventUpdate
				case database.EventDelete:
					op = EventDelete
				}

				select {
				case <-s.stream.Done():
					return
				case s.channel <- Event{OP: op, NodeID: id}:
				}
			}
		}
	}()

	return s
}

// Next returns a channel that receives Event notifications.
func (s *Stream) Next() <-chan Event {
	return s.channel
}

// Done returns a channel that is closed when the Stream is closed.
func (s *Stream) Done() <-chan struct{} {
	return s.stream.Done()
}

// Close closes the Stream.
func (s *Stream) Close() error {
	return s.stream.Close()
}
