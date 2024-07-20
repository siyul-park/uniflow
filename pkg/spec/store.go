package spec

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
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

const (
	EventStore  EventOP = iota // EventStore indicates an event for inserting a Spec.
	EventSwap                  // EventSwap indicates an event for updating a Spec.
	EventDelete                // EventDelete indicates an event for deleting a Spec.
)

// Common errors
var (
	ErrDuplicatedKey = errors.New("duplicated key") // ErrDuplicatedKey indicates a duplicated key error.
)
