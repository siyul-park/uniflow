package memdb

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/database"
)

// Stream manages a unidirectional stream of database events.
type Stream struct {
	in   chan database.Event
	out  chan database.Event
	done chan struct{}
	mu   sync.Mutex
}

var _ database.Stream = (*Stream)(nil)

func newStream() *Stream {
	s := &Stream{
		in:   make(chan database.Event),
		out:  make(chan database.Event),
		done: make(chan struct{}),
		mu:   sync.Mutex{},
	}

	go func() {
		defer close(s.out)
		buffer := make([]database.Event, 0, 4)

	loop:
		for {
			evt, ok := <-s.in
			if !ok {
				break loop
			}
			select {
			case s.out <- evt:
				continue
			default:
			}
			buffer = append(buffer, evt)
			for len(buffer) > 0 {
				select {
				case packet, ok := <-s.in:
					if !ok {
						break loop
					}
					buffer = append(buffer, packet)

				case s.out <- buffer[0]:
					buffer = buffer[1:]
				}
			}
		}
		for len(buffer) > 0 {
			s.out <- buffer[0]
			buffer = buffer[1:]
		}
	}()

	return s
}

// Next returns a receive-only channel for receiving events from the stream.
func (s *Stream) Next() <-chan database.Event {
	return s.out
}

// Done returns a receive-only channel that is closed when the stream is closed.
func (s *Stream) Done() <-chan struct{} {
	return s.done
}

// Close closes the stream, shutting down both input and signaling channels.
func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return nil
	default:
	}

	close(s.done)
	close(s.in)

	return nil
}

// Emit sends an event into the stream, if the stream is still open.
func (s *Stream) Emit(evt database.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
	default:
		s.in <- evt
	}
}
