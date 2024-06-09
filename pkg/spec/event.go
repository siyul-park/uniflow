package spec

import "github.com/gofrs/uuid"

// Event is an event that occurs when a spec.Spec is changed.
type Event struct {
	OP     eventOP
	NodeID uuid.UUID
}

type eventOP int

const (
	// EventInsert indicates an event for inserting a spec.Spec.
	EventInsert eventOP = iota
	// EventUpdate indicates an event for updating a spec.Spec.
	EventUpdate
	// EventDelete indicates an event for deleting a spec.Spec.
	EventDelete
)
