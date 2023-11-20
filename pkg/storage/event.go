package storage

import "github.com/oklog/ulid/v2"

type (
	// Event is an event that occurs when an scheme.Spec is changed.
	Event struct {
		OP     eventOP
		NodeID ulid.ULID
	}
	eventOP int
)

const (
	EventInsert eventOP = iota
	EventUpdate
	EventDelete
)
