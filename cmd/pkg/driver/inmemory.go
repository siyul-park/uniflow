package driver

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
)

// InMemoryDriver is a mock driver that provides in-memory stores.
type InMemoryDriver struct{}

var _ Driver = (*InMemoryDriver)(nil)

// NewInMemoryDriver creates and returns a new InMemoryDriver instance.
func NewInMemoryDriver() Driver {
	return &InMemoryDriver{}
}

// NewSpecStore creates and returns a new in-memory Spec Store.
func (c *InMemoryDriver) NewSpecStore(_ context.Context, _ string) (spec.Store, error) {
	return spec.NewStore(), nil
}

// NewValueStore creates and returns a new in-memory Value Store.
func (c *InMemoryDriver) NewValueStore(_ context.Context, _ string) (value.Store, error) {
	return value.NewStore(), nil
}

// Close is a no-op for InMemoryDriver, as there is no actual connection to close.
func (c *InMemoryDriver) Close(_ context.Context) error {
	return nil
}
