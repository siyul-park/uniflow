package database

import "context"

// Database is an interface for managing a database.
type Database interface {
	// Name returns the name of the database.
	Name() string

	// Collection returns a collection with the given name.
	Collection(ctx context.Context, name string) (Collection, error)

	// Drop deletes the entire database.
	Drop(ctx context.Context) error
}
