package database

import "context"

// Database is an abstracted interface for managing a database.
type Database interface {
	Name() string
	Collection(ctx context.Context, name string) (Collection, error)
	Drop(ctx context.Context) error
}
