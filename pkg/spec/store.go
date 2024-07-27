package spec

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// Store defines methods for managing Spec objects in a database.
type Store interface {
	// Watch returns a Stream that monitors changes matching the specified filter.
	Watch(ctx context.Context, specs ...Spec) (Stream, error)

	// Load retrieves Specs from the store that match the given criteria.
	Load(ctx context.Context, specs ...Spec) ([]Spec, error)

	// Store saves the given Specs into the database.
	Store(ctx context.Context, specs ...Spec) (int, error)

	// Swap updates existing Specs in the database with the provided data.
	Swap(ctx context.Context, specs ...Spec) (int, error)

	// Delete removes Specs from the store based on the provided criteria.
	Delete(ctx context.Context, specs ...Spec) (int, error)
}

// Stream represents a stream for tracking Spec changes.
type Stream interface {
	// Next returns a channel that receives Event notifications.
	Next() <-chan Event

	// Done returns a channel that is closed when the Stream is closed.
	Done() <-chan struct{}

	// Close closes the Stream.
	Close() error
}

// Event represents a change event for a Spec.
type Event struct {
	OP EventOP   // Operation type (Store, Swap, Delete)
	ID uuid.UUID // ID of the changed Spec
}

// EventOP represents the type of operation that triggered an Event.
type EventOP int

// store is an in-memory implementation of the Store interface using maps.
type store struct {
	data       map[uuid.UUID]Spec
	namespaces map[string]map[string]uuid.UUID
	streams    []*stream
	examples   [][]Spec
	mu         sync.RWMutex
}

// stream is an implementation of the Stream interface for memory streams.
type stream struct {
	in   chan Event
	out  chan Event
	done chan struct{}
	mu   sync.Mutex
}

const (
	EventStore  EventOP = iota // EventStore indicates an event for inserting a Spec.
	EventSwap                  // EventSwap indicates an event for updating a Spec.
	EventDelete                // EventDelete indicates an event for deleting a Spec.
)

// Common errors
var (
	ErrDuplicatedKey = errors.New("duplicated key") // ErrDuplicatedKey indicates a duplicated key error.
)

var _ Store = (*store)(nil)
var _ Stream = (*stream)(nil)

// NewStore creates a new MemStore instance.
func NewStore() Store {
	return &store{
		data:       make(map[uuid.UUID]Spec),
		namespaces: make(map[string]map[string]uuid.UUID),
	}
}

// Watch implements the Store interface, creating a stream for watching events.
func (s *store) Watch(ctx context.Context, specs ...Spec) (Stream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stream := newStream()

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

// Load implements the Store interface, loading specs matching the criteria.
func (s *store) Load(ctx context.Context, specs ...Spec) ([]Spec, error) {
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

// Store implements the Store interface, storing new specs.
func (s *store) Store(ctx context.Context, specs ...Spec) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, spec := range specs {
		if spec.GetID() == uuid.Nil {
			spec.SetID(uuid.Must(uuid.NewV7()))
		}

		if spec.GetNamespace() == "" {
			spec.SetNamespace(DefaultNamespace)
		}

		if spec.GetName() != "" && s.lookup(spec.GetNamespace(), spec.GetName()) != uuid.Nil {
			return 0, errors.WithStack(ErrDuplicatedKey)
		}
	}

	count := 0
	for _, spec := range specs {
		if s.insert(spec) {
			s.emit(EventStore, spec)
			count++
		}
	}
	return count, nil
}

// Swap implements the Store interface, swapping existing specs with new ones.
func (s *store) Swap(ctx context.Context, specs ...Spec) (int, error) {
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

	for i := 0; i < len(specs); i++ {
		spec := specs[i]
		if !s.free(spec.GetID()) {
			specs = append(specs[:i], specs[i+1:]...)
			i--
		}
	}

	count := 0
	for _, spec := range specs {
		if !s.insert(spec) {
			return 0, errors.WithStack(ErrDuplicatedKey)
		}
		s.emit(EventSwap, spec)
		count++
	}
	return count, nil
}

// Delete implements the Store interface, deleting specs matching the criteria.
func (s *store) Delete(ctx context.Context, specs ...Spec) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, spec := range s.data {
		if s.match(spec, specs...) {
			if s.free(id) {
				s.emit(EventDelete, spec)
				count++
			}
		}
	}
	return count, nil
}

func (s *store) match(spec Spec, examples ...Spec) bool {
	for i, example := range examples {
		if example == nil {
			examples = append(examples[:i], examples[i+1:]...)
			i--
		}
	}

	if len(examples) == 0 {
		return true
	}
	return len(Match(spec, examples...)) > 0
}

func (s *store) insert(spec Spec) bool {
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

func (s *store) free(id uuid.UUID) bool {
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

func (s *store) lookup(namespace, name string) uuid.UUID {
	if ns, ok := s.namespaces[namespace]; ok {
		return ns[name]
	}
	return uuid.Nil
}

func (s *store) emit(op EventOP, spec Spec) {
	for i, stream := range s.streams {
		if s.match(spec, s.examples[i]...) {
			stream.Emit(Event{
				OP: op,
				ID: spec.GetID(),
			})
		}
	}
}

// newStream creates a new memory stream for event notifications.
func newStream() *stream {
	s := &stream{
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
func (s *stream) Next() <-chan Event {
	return s.out
}

// Done returns a receive-only channel that is closed when the stream is closed.
func (s *stream) Done() <-chan struct{} {
	return s.done
}

// Close closes the stream, shutting down both input and signaling channels.
func (s *stream) Close() error {
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
func (s *stream) Emit(event Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
	default:
		s.in <- event
	}
}
