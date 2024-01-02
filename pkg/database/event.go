package database

import "github.com/siyul-park/uniflow/pkg/primitive"

// Event represents an event that occurred in the collection.
type Event struct {
	OP         EventOP
	DocumentID primitive.Value
}

// EventOP represents the type of operation in a collection event.
type EventOP int

const (
	EventInsert EventOP = iota
	EventUpdate
	EventDelete
)
