package spec

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// MemStore is an in-memory implementation of the Store interface using maps.
type MemStore struct {
	data       map[uuid.UUID]Spec
	namespaces map[string]map[string]uuid.UUID
	streams    []*MemStream
	examples   [][]Spec
	mu         sync.RWMutex
}

// MemStream is an implementation of the Stream interface for memory streams.
type MemStream struct {
	in   chan Event
	out  chan Event
	done chan struct{}
	mu   sync.Mutex
}

var _ Store = (*MemStore)(nil)
var _ Stream = (*MemStream)(nil)

// NewMemStore creates a new MemStore instance.
func NewMemStore() *MemStore {
	return &MemStore{
		data:       make(map[uuid.UUID]Spec),
		namespaces: make(map[string]map[string]uuid.UUID),
	}
}

// Watch implements the Store interface.
func (s *MemStore) Watch(ctx context.Context, specs ...Spec) (Stream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stream := newMemStream()

	go func() {
		select {
		case <-ctx.Done():
			stream.Close()
		case <-stream.Done():
		}
	}()

	go func() {
		<-stream.Done()

		s.mu.Lock()
		defer s.mu.Unlock()

		for i, it := range s.streams {
			if it == stream {
				s.streams = append(s.streams[:i], s.streams[i+1:]...)
				s.examples = append(s.examples[:i], s.examples[i+1:]...)
				break
			}
		}
	}()

	s.streams = append(s.streams, stream)
	s.examples = append(s.examples, specs)

	return stream, nil
}

// Load implements the Store interface.
func (s *MemStore) Load(ctx context.Context, specs ...Spec) ([]Spec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Spec
	for _, spec := range s.data {
		if s.match(spec, specs...) {
			result = append(result, spec)
		}
	}
	return result, nil
}

// Store implements the Store interface.
func (s *MemStore) Store(ctx context.Context, specs ...Spec) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, spec := range specs {
		if spec.GetNamespace() == "" {
			spec.SetNamespace(DefaultNamespace)
		}

		if spec.GetID() == uuid.Nil {
			spec.SetID(uuid.Must(uuid.NewV7()))
		}

		if spec.GetName() != "" && s.lookup(spec.GetNamespace(), spec.GetName()) != uuid.Nil {
			return 0, errors.WithStack(ErrDuplicatedKey)
		}
	}

	count := 0
	for _, spec := range specs {
		if s.insert(spec) {
			for i, stream := range s.streams {
				if s.match(spec, s.examples[i]...) {
					stream.Emit(Event{
						OP: EventStore,
						ID: spec.GetID(),
					})
				}
			}
			count++
		}
	}
	return count, nil
}

// Swap implements the Store interface.
func (s *MemStore) Swap(ctx context.Context, specs ...Spec) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, spec := range specs {
		if spec.GetNamespace() == "" {
			spec.SetNamespace(DefaultNamespace)
		}

		if spec.GetID() == uuid.Nil {
			spec.SetID(s.lookup(spec.GetNamespace(), spec.GetName()))
		}
	}

	count := 0
	for _, spec := range specs {
		if s.free(spec.GetID()) && s.insert(spec) {
			for i, stream := range s.streams {
				if s.match(spec, s.examples[i]...) {
					stream.Emit(Event{
						OP: EventSwap,
						ID: spec.GetID(),
					})
				}
			}
			count++
		}
	}
	return count, nil
}

// Delete implements the Store interface.
func (s *MemStore) Delete(ctx context.Context, specs ...Spec) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, spec := range s.data {
		if s.match(spec, specs...) {
			if s.free(id) {
				for i, stream := range s.streams {
					if s.match(spec, s.examples[i]...) {
						stream.Emit(Event{
							OP: EventDelete,
							ID: spec.GetID(),
						})
					}
				}
				count++
			}
		}
	}
	return count, nil
}

func (s *MemStore) match(spec Spec, examples ...Spec) bool {
	if len(examples) == 0 {
		return true
	}

	for _, example := range examples {
		if example == nil ||
			(example.GetID() != uuid.Nil && spec.GetID() != example.GetID()) ||
			(example.GetNamespace() != "" && spec.GetNamespace() != example.GetNamespace()) ||
			(example.GetName() != "" && spec.GetName() != example.GetName()) {
			continue
		}
		return true
	}
	return false
}

func (s *MemStore) insert(spec Spec) bool {
	if _, exists := s.data[spec.GetID()]; exists {
		return false
	}

	id := s.lookup(spec.GetNamespace(), spec.GetName())
	if id != uuid.Nil && id != spec.GetID() {
		return false
	}

	s.data[spec.GetID()] = spec

	if spec.GetName() != "" {
		ns, ok := s.namespaces[spec.GetNamespace()]
		if !ok {
			ns = make(map[string]uuid.UUID)
			s.namespaces[spec.GetNamespace()] = ns
		}
		ns[spec.GetName()] = spec.GetID()
	}
	return true
}

func (s *MemStore) free(id uuid.UUID) bool {
	spec, ok := s.data[id]
	if !ok {
		return false
	}

	if spec.GetName() != "" {
		if ns, ok := s.namespaces[spec.GetNamespace()]; ok {
			delete(ns, spec.GetName())
			if len(ns) == 0 {
				delete(s.namespaces, spec.GetNamespace())
			}
		}
	}
	delete(s.data, id)
	return true
}

func (s *MemStore) lookup(namespace, name string) uuid.UUID {
	if ns, ok := s.namespaces[namespace]; ok {
		return ns[name]
	}
	return uuid.Nil
}

func newMemStream() *MemStream {
	s := &MemStream{
		in:   make(chan Event),
		out:  make(chan Event),
		done: make(chan struct{}),
	}

	go func() {
		defer close(s.out)
		defer close(s.in)

		buffer := make([]Event, 0, 2)
		for {
			var event Event
			select {
			case event = <-s.in:
			case <-s.done:
				return
			}

			select {
			case s.out <- event:
			default:
				buffer = append(buffer, event)

				for len(buffer) > 0 {
					select {
					case event = <-s.in:
						buffer = append(buffer, event)
					case s.out <- buffer[0]:
						buffer = buffer[1:]
					}
				}
			}
		}
	}()

	return s
}

// Next returns a receive-only channel for receiving events from the stream.
func (s *MemStream) Next() <-chan Event {
	return s.out
}

// Done returns a receive-only channel that is closed when the stream is closed.
func (s *MemStream) Done() <-chan struct{} {
	return s.done
}

// Close closes the stream, shutting down both input and signaling channels.
func (s *MemStream) Close() error {
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

// Emit sends an event into the stream, if the stream is still open.
func (s *MemStream) Emit(event Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
	default:
		s.in <- event
	}
}
