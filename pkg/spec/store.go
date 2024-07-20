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

// Event represents an event that occurs when a Spec is changed.
type Event struct {
	OP EventOP
	ID uuid.UUID
}

type EventOP int

const (
	// EventStore indicates an event for inserting a Spec.
	EventStore EventOP = iota
	// EventSwap indicates an event for updating a Spec.
	EventSwap
	// EventDelete indicates an event for deleting a Spec.
	EventDelete
)

var (
	// ErrDuplicatedKey indicates a duplicated key error.
	ErrDuplicatedKey = errors.New("duplicated key")
)
