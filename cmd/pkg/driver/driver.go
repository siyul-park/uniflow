package driver

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/store"
)

// Driver defines an interface for interacting with various storage mechanisms.
type Driver interface {
	// NewSpecStore returns a store for managing specifications.
	NewSpecStore(ctx context.Context, name string) (store.Store, error)

	// NewValueStore returns a store for managing values.
	NewValueStore(ctx context.Context, name string) (store.Store, error)

	// Close terminates the connection and releases resources.
	Close(ctx context.Context) error
}
