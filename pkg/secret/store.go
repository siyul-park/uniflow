package secret

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// Store defines methods for managing Secret objects in a database.
type Store interface {
	// Watch returns a Stream that monitors changes matching the specified filter.
	Watch(ctx context.Context, secrets ...Secret) (Stream, error)

	// Load retrieves Secrets from the store that match the given criteria.
	Load(ctx context.Context, secrets ...Secret) ([]Secret, error)

	// Store saves the given Secrets into the database.
	Store(ctx context.Context, secrets ...Secret) (int, error)

	// Swap updates existing Secrets in the database with the provided data.
	Swap(ctx context.Context, secrets ...Secret) (int, error)

	// Delete removes Secrets from the store based on the provided criteria.
	Delete(ctx context.Context, secrets ...Secret) (int, error)
}

// Stream represents a stream for tracking Secret changes.
type Stream interface {
	// Next returns a channel that receives Event notifications.
	Next() <-chan Event

	// Done returns a channel that is closed when the Stream is closed.
	Done() <-chan struct{}

	// Close closes the Stream.
	Close() error
}

// Event represents a change event for a Secret.
type Event struct {
	OP EventOP   // Operation type (Store, Swap, Delete)
	ID uuid.UUID // ID of the changed Secret
}

// EventOP represents the type of operation that triggered an Event.
type EventOP int

const (
	EventStore  EventOP = iota // EventStore indicates an event for inserting a Secret.
	EventSwap                  // EventSwap indicates an event for updating a Secret.
	EventDelete                // EventDelete indicates an event for deleting a Secret.
)

// store is an in-memory implementation of the Store interface using maps.
type store struct {
	data       map[uuid.UUID]Secret
	namespaces map[string]map[string]uuid.UUID
	streams    []*stream
	examples   [][]Secret
	mu         sync.RWMutex
}

// stream is an implementation of the Stream interface for memory streams.
type stream struct {
	in   chan Event
	out  chan Event
	done chan struct{}
	mu   sync.Mutex
}

var (
	// Common errors
	ErrDuplicatedKey = errors.New("duplicated key") // ErrDuplicatedKey indicates a duplicated key error.
)

var _ Store = (*store)(nil)
var _ Stream = (*stream)(nil)

// NewStore creates a new Store instance for managing Secrets.
func NewStore() Store {
	return &store{
		data:       make(map[uuid.UUID]Secret),
		namespaces: make(map[string]map[string]uuid.UUID),
	}
}

// Watch implements the Store interface, creating a stream for watching events.
func (s *store) Watch(ctx context.Context, secrets ...Secret) (Stream, error) {
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
	s.examples = append(s.examples, secrets)

	return stream, nil
}

// Load implements the Store interface, loading secrets matching the criteria.
func (s *store) Load(ctx context.Context, secrets ...Secret) ([]Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Secret
	for _, secret := range s.data {
		if s.match(secret, secrets...) {
			result = append(result, secret)
		}
	}
	return result, nil
}

// Store implements the Store interface, storing new secrets.
func (s *store) Store(ctx context.Context, secrets ...Secret) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, secret := range secrets {
		if secret.GetID() == uuid.Nil {
			secret.SetID(uuid.Must(uuid.NewV7()))
		}

		if secret.GetNamespace() == "" {
			secret.SetNamespace(DefaultNamespace)
		}

		if secret.GetName() != "" && s.lookup(secret.GetNamespace(), secret.GetName()) != uuid.Nil {
			return 0, errors.WithStack(ErrDuplicatedKey)
		}
	}

	count := 0
	for _, secret := range secrets {
		if s.insert(secret) {
			s.emit(EventStore, secret)
			count++
		}
	}
	return count, nil
}

// Swap implements the Store interface, swapping existing secrets with new ones.
func (s *store) Swap(ctx context.Context, secrets ...Secret) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, secret := range secrets {
		if secret.GetNamespace() == "" {
			secret.SetNamespace(DefaultNamespace)
		}

		if secret.GetID() == uuid.Nil {
			secret.SetID(s.lookup(secret.GetNamespace(), secret.GetName()))
		}
	}

	count := 0
	for _, secret := range secrets {
		if s.free(secret.GetID()) && s.insert(secret) {
			s.emit(EventSwap, secret)
			count++
		}
	}
	return count, nil
}

// Delete implements the Store interface, deleting secrets matching the criteria.
func (s *store) Delete(ctx context.Context, secrets ...Secret) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, secret := range s.data {
		if s.match(secret, secrets...) {
			if s.free(id) {
				s.emit(EventDelete, secret)
				count++
			}
		}
	}
	return count, nil
}

func (s *store) match(secret Secret, examples ...Secret) bool {
	for i, example := range examples {
		if example == nil {
			examples = append(examples[:i], examples[i+1:]...)
			i--
		}
	}

	if len(examples) == 0 {
		return true
	}

	for _, example := range examples {
		if example.GetID() != uuid.Nil && secret.GetID() != example.GetID() {
			continue
		}
		if example.GetNamespace() != "" && secret.GetNamespace() != example.GetNamespace() {
			continue
		}
		if example.GetName() != "" && secret.GetName() != example.GetName() {
			continue
		}
		return true
	}
	return false
}

func (s *store) insert(secret Secret) bool {
	if _, exists := s.data[secret.GetID()]; exists {
		return false
	}

	id := s.lookup(secret.GetNamespace(), secret.GetName())
	if id != uuid.Nil && id != secret.GetID() {
		return false
	}

	s.data[secret.GetID()] = secret

	if secret.GetName() != "" {
		ns, ok := s.namespaces[secret.GetNamespace()]
		if !ok {
			ns = make(map[string]uuid.UUID)
			s.namespaces[secret.GetNamespace()] = ns
		}
		ns[secret.GetName()] = secret.GetID()
	}
	return true
}

func (s *store) free(id uuid.UUID) bool {
	secret, ok := s.data[id]
	if !ok {
		return false
	}

	if secret.GetName() != "" {
		if ns, ok := s.namespaces[secret.GetNamespace()]; ok {
			delete(ns, secret.GetName())
			if len(ns) == 0 {
				delete(s.namespaces, secret.GetNamespace())
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

func (s *store) emit(op EventOP, secret Secret) {
	for i, stream := range s.streams {
		if s.match(secret, s.examples[i]...) {
			stream.Emit(Event{
				OP: op,
				ID: secret.GetID(),
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
