package event

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/primitive"
)

type Event struct {
	data map[string]primitive.Value
	mu   sync.RWMutex
}

const KeyTopic = "topic"

var _ primitive.Marshaler = (*Event)(nil)
var _ primitive.Unmarshaler = (*Event)(nil)

func New(topic string) *Event {
	return &Event{
		data: map[string]primitive.Value{
			KeyTopic: primitive.NewString(topic),
		},
	}
}

func (e *Event) Topic() string {
	if v, ok := e.Get(KeyTopic); ok {
		if v, ok := v.(primitive.String); ok {
			return v.String()
		}
	}
	return ""
}

func (e *Event) Set(key string, val primitive.Value) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.data[key] = val
}

func (e *Event) Get(key string) (primitive.Value, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	val, ok := e.data[key]
	return val, ok
}

func (e *Event) MarshalPrimitive() (primitive.Value, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return primitive.MarshalBinary(e.data)
}

func (e *Event) UnmarshalPrimitive(value primitive.Value) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return primitive.Unmarshal(value, &e.data)
}
