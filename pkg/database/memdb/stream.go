package memdb

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/database"
)

type (
	Stream struct {
		buffer  []database.Event
		channel chan database.Event
		pump    chan struct{}
		done    chan struct{}
		mu      sync.Mutex
	}
)

func NewStream() *Stream {
	s := &Stream{
		buffer:  nil,
		channel: make(chan database.Event),
		pump:    make(chan struct{}),
		done:    make(chan struct{}),
		mu:      sync.Mutex{},
	}

	go func() {
		defer func() { close(s.channel) }()

		for {
			select {
			case <-s.done:
				return
			case <-s.pump:
				buffer := func() []database.Event {
					s.mu.Lock()
					defer s.mu.Unlock()

					buffer := s.buffer
					s.buffer = nil
					return buffer
				}()

				for _, event := range buffer {
					select {
					case <-s.done:
						return
					case s.channel <- event:
					}
				}
			}
		}
	}()

	return s
}

func (s *Stream) Next() <-chan database.Event {
	return s.channel
}

func (s *Stream) Done() <-chan struct{} {
	return s.done
}

func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return nil
	default:
	}

	close(s.done)
	s.buffer = nil

	return nil
}

func (s *Stream) Emit(event database.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
	default:
		s.buffer = append(s.buffer, event)
		s.push()
	}
}

func (p *Stream) push() {
	go func() {
		select {
		case <-p.done:
		default:
			p.pump <- struct{}{}
		}
	}()
}
