package driver

import (
	"context"
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// InMemoryDriver is a mock driver that provides in-memory stores.
type InMemoryDriver struct{}

var _ Driver = (*InMemoryDriver)(nil)

// NewInMemoryDriver creates and returns a new InMemoryDriver instance.
func NewInMemoryDriver() Driver {
	return &InMemoryDriver{}
}

// SpecStore creates and returns a new in-memory Spec Store.
func (c *InMemoryDriver) SpecStore(_ context.Context, _ string) (spec.Store, error) {
	return spec.NewStore(), nil
}

// SecretStore creates and returns a new in-memory Secret Store.
func (c *InMemoryDriver) SecretStore(_ context.Context, _ string) (secret.Store, error) {
	return secret.NewStore(), nil
}

// ChartStore creates and returns a new in-memory Chart Store.
func (c *InMemoryDriver) ChartStore(_ context.Context, _ string) (chart.Store, error) {
	return chart.NewStore(), nil
}

// Close is a no-op for InMemoryDriver, as there is no actual connection to close.
func (c *InMemoryDriver) Close(_ context.Context) error {
	return nil
}
