package driver

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
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
func (c *InMemoryDriver) NewSpecStore(ctx context.Context, _ string) (store.Store, error) {
	s := store.New()
	err := s.Index(ctx, []string{spec.KeyNamespace, spec.KeyName}, store.IndexOptions{
		Unique: true,
		Filter: map[string]any{"name": map[string]any{"$exists": true}},
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

// NewValueStore creates and returns a new in-memory Value Store.
func (c *InMemoryDriver) NewValueStore(ctx context.Context, _ string) (store.Store, error) {
	s := store.New()
	err := s.Index(ctx, []string{value.KeyNamespace, value.KeyName}, store.IndexOptions{
		Unique: true,
		Filter: map[string]any{"name": map[string]any{"$exists": true}},
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Close is a no-op for InMemoryDriver, as there is no actual connection to close.
func (c *InMemoryDriver) Close(_ context.Context) error {
	return nil
}
