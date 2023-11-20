package database

import "context"

type (
	// IndexView is an abstracted interface for be used to create, drop, and list indexes.
	IndexView interface {
		List(ctx context.Context) ([]IndexModel, error)
		Create(ctx context.Context, index IndexModel) error
		Drop(ctx context.Context, name string) error
	}

	// IndexModel is a model for an index.
	IndexModel struct {
		Name    string
		Keys    []string
		Unique  bool
		Partial *Filter
	}
)
