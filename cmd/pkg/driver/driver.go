package driver

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
)

// Driver defines an interface for interacting with various storage mechanisms.
type Driver interface {
	// SpecStore returns a store for managing specifications.
	SpecStore(ctx context.Context, name string) (spec.Store, error)

	// ValueStore returns a store for managing values.
	ValueStore(ctx context.Context, name string) (value.Store, error)

	// ChartStore returns a store for managing charts.
	ChartStore(ctx context.Context, name string) (chart.Store, error)

	// Close terminates the connection and releases resources.
	Close(ctx context.Context) error
}
