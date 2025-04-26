package driver

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"

	"github.com/siyul-park/uniflow/pkg/types"
)

// Stream defines an interface for consuming event streams.
type Stream interface {
	Next(ctx context.Context) bool
	Decode(val any) error
	Close(ctx context.Context) error
}

// Event represents a change event within the store.
type Event struct {
	ID uuid.UUID `json:"id" yaml:"id" validate:"required"`
	OP string    `json:"op" yaml:"op" validate:"required"`
}

type stream struct {
	doc  types.Map
	in   chan types.Map
	out  chan types.Map
	done chan struct{}
	mu   sync.Mutex
}

var _ Stream = (*stream)(nil)

func newStream() *stream {
	c := &stream{
		in:   make(chan types.Map),
		out:  make(chan types.Map),
		done: make(chan struct{}),
	}

	go func() {
		defer close(c.out)
		defer close(c.in)

		buffer := make([]types.Map, 0, 2)
		for {
			var event types.Map
			select {
			case event = <-c.in:
			case <-c.done:
				return
			}

			select {
			case c.out <- event:
			case <-c.done:
				return
			default:
				buffer = append(buffer, event)

				for len(buffer) > 0 {
					select {
					case event = <-c.in:
						buffer = append(buffer, event)
					case c.out <- buffer[0]:
						buffer = buffer[1:]
					case <-c.done:
						return
					}
				}
			}
		}
	}()

	return c
}

func (s *stream) Emit(doc types.Map) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return false
	default:
		s.in <- doc
		return true
	}
}

func (s *stream) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case doc, ok := <-s.out:
		s.doc = doc
		return ok
	}
}

func (s *stream) Decode(val any) error {
	return types.Unmarshal(s.doc, val)
}

func (s *stream) Close(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return nil
	default:
		close(s.done)
		return nil
	}
}

func (s *stream) Done() <-chan struct{} {
	return s.done
}
