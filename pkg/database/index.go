package database

import "context"

// IndexView is an abstracted interface used to create, drop, and list indexes.
type IndexView interface {
	List(ctx context.Context) ([]IndexModel, error)
	Create(ctx context.Context, index IndexModel) error
	Drop(ctx context.Context, name string) error
}

// IndexModel represents a model for an index.
type IndexModel struct {
	Name    string
	Keys    []string
	Unique  bool
	Partial *Filter
}
