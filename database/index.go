package database

import "context"

// IndexView is an abstracted interface used to manage indexes.
type IndexView interface {
	List(ctx context.Context) ([]IndexModel, error)     // List returns a list of indexes.
	Create(ctx context.Context, index IndexModel) error // Create creates a new index.
	Drop(ctx context.Context, name string) error        // Drop deletes an index by name.
}

// IndexModel represents a model for an index.
type IndexModel struct {
	Name    string   // Name represents the name of the index.
	Keys    []string // Keys represents the fields included in the index.
	Unique  bool     // Unique indicates if the index enforces uniqueness.
	Partial *Filter  // Partial represents a filter for a partial index.
}
