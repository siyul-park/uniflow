package database

import "context"

type (
	// Database is an abstracted interface for managing database.
	Database interface {
		Name() string
		Collection(ctx context.Context, name string) (Collection, error)
		Drop(ctx context.Context) error
	}
)
