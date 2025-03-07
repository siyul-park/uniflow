package resource

import (
	"context"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Store defines methods for managing Resource objects in a database.
type Store[T Resource] interface {
	// Watch returns a Stream that monitors changes matching the specified filter.
	Watch(ctx context.Context, resources ...T) (Stream, error)

	// Load retrieves resources from the store that match the given criteria.
	Load(ctx context.Context, resources ...T) ([]T, error)

	// Store saves the given resources into the database.
	Store(ctx context.Context, resources ...T) (int, error)

	// Swap updates existing resources in the database with the provided data.
	Swap(ctx context.Context, resources ...T) (int, error)

	// Delete removes resources from the store based on the provided criteria.
	Delete(ctx context.Context, resources ...T) (int, error)
}

// Stream represents a stream for tracking Resource changes.
type Stream interface {
	// Next returns a channel that receives Event notifications.
	Next() <-chan Event

	// Done returns a channel that is closed when the Stream is closed.
	Done() <-chan struct{}

	// Close closes the Stream.
	Close() error
}

// Event represents a change event for a Resource.
type Event struct {
	ID uuid.UUID `json:"id"`
	OP EventOP   `json:"op"`
}

// EventOP represents the type of operation that triggered an Event.
type EventOP string

type store[T Resource] struct {
	data       map[uuid.UUID]T
	namespaces map[string]map[string]uuid.UUID
	streams    []*stream
	examples   [][]T
	validate   *validator.Validate
	mu         sync.RWMutex
}

type stream struct {
	in   chan Event
	out  chan Event
	done chan struct{}
	mu   sync.Mutex
}

const (
	EventStore  EventOP = "store"
	EventSwap   EventOP = "swap"
	EventDelete EventOP = "delete"
)

var ErrDuplicatedKey = errors.New("duplicated key") // ErrDuplicatedKey indicates a duplicated key error.

var _ Store[Resource] = (*store[Resource])(nil)
var _ Stream = (*stream)(nil)

// NewStore creates a new store instance.
func NewStore[T Resource]() Store[T] {
	return &store[T]{
		data:       make(map[uuid.UUID]T),
		namespaces: make(map[string]map[string]uuid.UUID),
		validate:   validator.New(validator.WithRequiredStructEnabled()),
	}
}

// Watch implements the Store interface, creating a stream for watching events.
func (s *store[T]) Watch(ctx context.Context, resources ...T) (Stream, error) {
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
	s.examples = append(s.examples, resources)

	return stream, nil
}

// Load implements the Store interface, loading resources matching the criteria.
func (s *store[T]) Load(_ context.Context, resources ...T) ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []T
	for _, res := range s.data {
		if s.match(res, resources...) {
			result = append(result, res)
		}
	}
	return result, nil
}

// Store implements the Store interface, storing new resources.
func (s *store[T]) Store(_ context.Context, resources ...T) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, res := range resources {
		if res.GetID() == uuid.Nil {
			res.SetID(uuid.Must(uuid.NewV7()))
		}

		if res.GetNamespace() == "" {
			res.SetNamespace(DefaultNamespace)
		}

		if err := s.validate.Struct(res); err != nil {
			return 0, errors.WithMessage(encoding.ErrUnsupportedValue, err.Error())
		}

		if res.GetName() != "" && s.lookup(res.GetNamespace(), res.GetName()) != uuid.Nil {
			return 0, errors.WithStack(ErrDuplicatedKey)
		}
	}

	count := 0
	for _, res := range resources {
		if s.insert(res) {
			s.emit(EventStore, res)
			count++
		}
	}
	return count, nil
}

// Swap implements the Store interface, swapping existing resources with new ones.
func (s *store[T]) Swap(_ context.Context, resources ...T) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, res := range resources {
		if res.GetNamespace() == "" {
			res.SetNamespace(DefaultNamespace)
		}

		if res.GetID() == uuid.Nil {
			res.SetID(s.lookup(res.GetNamespace(), res.GetName()))
		}

		if err := s.validate.Struct(res); err != nil {
			return 0, errors.WithMessage(encoding.ErrUnsupportedValue, err.Error())
		}
	}

	for i := 0; i < len(resources); i++ {
		res := resources[i]
		if !s.free(res.GetID()) {
			resources = append(resources[:i], resources[i+1:]...)
			i--
		}
	}

	count := 0
	for _, res := range resources {
		if !s.insert(res) {
			return 0, errors.WithStack(ErrDuplicatedKey)
		}
		s.emit(EventSwap, res)
		count++
	}
	return count, nil
}

// Delete implements the Store interface, deleting resources matching the criteria.
func (s *store[T]) Delete(_ context.Context, resources ...T) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, res := range s.data {
		if s.match(res, resources...) {
			if s.free(id) {
				s.emit(EventDelete, res)
				count++
			}
		}
	}
	return count, nil
}

func (s *store[T]) match(source T, targets ...T) bool {
	if len(targets) == 0 {
		return true
	}
	for _, target := range targets {
		if Is(source, target) {
			return true
		}
	}
	return false
}

func (s *store[T]) insert(res T) bool {
	if _, ok := s.data[res.GetID()]; ok {
		return false
	}

	id := s.lookup(res.GetNamespace(), res.GetName())
	if id != uuid.Nil && id != res.GetID() {
		return false
	}

	s.data[res.GetID()] = res

	if res.GetName() != "" {
		ns, ok := s.namespaces[res.GetNamespace()]
		if !ok {
			ns = make(map[string]uuid.UUID)
			s.namespaces[res.GetNamespace()] = ns
		}
		ns[res.GetName()] = res.GetID()
	}
	return true
}

func (s *store[T]) free(id uuid.UUID) bool {
	res, ok := s.data[id]
	if !ok {
		return false
	}

	if res.GetName() != "" {
		if ns, ok := s.namespaces[res.GetNamespace()]; ok {
			delete(ns, res.GetName())
			if len(ns) == 0 {
				delete(s.namespaces, res.GetNamespace())
			}
		}
	}
	delete(s.data, id)
	return true
}

func (s *store[T]) lookup(namespace, name string) uuid.UUID {
	if ns, ok := s.namespaces[namespace]; ok {
		return ns[name]
	}
	return uuid.Nil
}

func (s *store[T]) emit(op EventOP, resource T) {
	for i, stream := range s.streams {
		if s.match(resource, s.examples[i]...) {
			stream.Emit(Event{
				OP: op,
				ID: resource.GetID(),
			})
		}
	}
}

// newStream creates a new in-memory stream for event notifications.
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
