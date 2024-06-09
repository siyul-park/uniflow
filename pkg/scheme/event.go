package scheme

import "github.com/gofrs/uuid"

// Event is an event that occurs when a scheme.Spec is changed.
type Event struct {
	OP     eventOP
	NodeID uuid.UUID
}

type eventOP int

const (
	// EventInsert indicates an event for inserting a scheme.Spec.
	EventInsert eventOP = iota
	// EventUpdate indicates an event for updating a scheme.Spec.
	EventUpdate
	// EventDelete indicates an event for deleting a scheme.Spec.
	EventDelete
)
