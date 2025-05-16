package plugin

import (
	"context"
	"github.com/pkg/errors"
)

// Plugin defines the interface that dynamic plugins must implement.
type Plugin interface {
	Name() string
	Version() string
	Load(ctx context.Context) error
	Unload(ctx context.Context) error
}

var ErrMissingDependency = errors.New("missing dependency")
