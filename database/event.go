package database

import "github.com/siyul-park/uniflow/object"

// Event represents an event that occurred in the collection.
type Event struct {
	OP         EventOP       // Type of operation in the collection event.
	DocumentID object.Object // ID of the document associated with the event.
}

// EventOP represents the type of operation in a collection event.
type EventOP int

const (
	EventInsert EventOP = iota // Insert operation event.
	EventUpdate                // Update operation event.
	EventDelete                // Delete operation event.
)
