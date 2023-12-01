package storage

import "github.com/oklog/ulid/v2"

type (
	// Event is an event that occurs when a scheme.Spec is changed.
	Event struct {
		OP     eventOP
		NodeID ulid.ULID
	}
	eventOP int
)

const (
	// EventInsert indicates an event for inserting a scheme.Spec.
	EventInsert eventOP = iota
	// EventUpdate indicates an event for updating a scheme.Spec.
	EventUpdate
	// EventDelete indicates an event for deleting a scheme.Spec.
	EventDelete
)
