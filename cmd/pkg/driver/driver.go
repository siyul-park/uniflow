package driver

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
)

// Driver defines an interface for interacting with various storage mechanisms.
type Driver interface {
	// NewSpecStore returns a store for managing specifications.
	NewSpecStore(ctx context.Context, name string) (spec.Store, error)

	// NewValueStore returns a store for managing values.
	NewValueStore(ctx context.Context, name string) (value.Store, error)

	// Close terminates the connection and releases resources.
	Close(ctx context.Context) error
}
