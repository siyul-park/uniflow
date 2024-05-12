package event

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/primitive"
)

// Event represents a structured event with associated data.
type Event struct {
	data map[string]primitive.Value
	mu   sync.RWMutex
}

const KeyTopic = "topic"

var _ primitive.Marshaler = (*Event)(nil)
var _ primitive.Unmarshaler = (*Event)(nil)

// New creates a new Event instance with the specified topic.
func New(topic string) *Event {
	return &Event{
		data: map[string]primitive.Value{
			KeyTopic: primitive.NewString(topic),
		},
	}
}

// Topic returns the topic of the event.
func (e *Event) Topic() string {
	if v, ok := e.Get(KeyTopic); ok {
		if v, ok := v.(primitive.String); ok {
			return v.String()
		}
	}
	return ""
}

// Set sets the value associated with the specified key in the event data.
func (e *Event) Set(key string, val primitive.Value) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.data[key] = val
}

// Get retrieves the value associated with the specified key from the event data.
func (e *Event) Get(key string) (primitive.Value, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	val, ok := e.data[key]
	return val, ok
}

// MarshalPrimitive marshals the event data to a primitive value.
func (e *Event) MarshalPrimitive() (primitive.Value, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return primitive.MarshalBinary(e.data)
}

// UnmarshalPrimitive unmarshals the primitive value to populate the event data.
func (e *Event) UnmarshalPrimitive(value primitive.Value) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return primitive.Unmarshal(value, &e.data)
}
