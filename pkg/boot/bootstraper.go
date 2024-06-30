package boot

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/loader"
)

// BootstraperConfig holds configuration for initializing a Bootstraper.
type BootstraperConfig struct {
	Loader     *loader.Loader     
	Reconciler *loader.Reconciler 
	BootHooks  []BootHook         
}

// Bootstraper manages resource loading and executes boot hooks during initialization.
type Bootstraper struct {
	loader     *loader.Loader
	reconciler *loader.Reconciler
	bootHooks  []BootHook
}

// NewBootstraper creates a new Bootstraper with the given configuration.
func NewBootstraper(config BootstraperConfig) *Bootstraper {
	return &Bootstraper{
		loader:     config.Loader,
		reconciler: config.Reconciler,
		bootHooks:  config.BootHooks,
	}
}

// Boot synchronizes and loads resources upon initialization.
func (b *Bootstraper) Boot(ctx context.Context) error {
	if err := b.reconciler.Watch(ctx); err != nil {
		return err
	}

	if _, err := b.loader.LoadAll(ctx); err != nil {
		return err
	}

	for _, hook := range b.bootHooks {
		if err := hook.Boot(ctx); err != nil {
			return err
		}
	}

	return nil
}
