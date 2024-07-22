package runtime

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Config defines configuration options for the Runtime.
type Config struct {
	Namespace string         // Namespace defines the isolated execution environment for workflows.
	Hook      *hook.Hook     // Hook is a collection of hook functions for managing symbols.
	Scheme    *scheme.Scheme // Scheme defines the scheme and behaviors for symbols.
	Store     spec.Store     // Store is responsible for persisting symbols.
}

// Runtime represents an environment for executing Workflows.
type Runtime struct {
	namespace string
	scheme    *scheme.Scheme
	store     spec.Store
	table     *symbol.Table
	loader    *symbol.Loader
}

// New creates a new Runtime instance with the specified configuration.
func New(config Config) *Runtime {
	if config.Namespace == "" {
		config.Namespace = spec.DefaultNamespace
	}
	if config.Hook == nil {
		config.Hook = hook.New()
	}
	if config.Scheme == nil {
		config.Scheme = scheme.New()
	}
	if config.Store == nil {
		config.Store = spec.NewStore()
	}

	tb := symbol.NewTable(symbol.TableOptions{
		LoadHooks:   []symbol.LoadHook{config.Hook},
		UnloadHooks: []symbol.UnloadHook{config.Hook},
	})

	ld := symbol.NewLoader(symbol.LoaderConfig{
		Scheme: config.Scheme,
		Store:  config.Store,
		Table:  tb,
	})

	return &Runtime{
		namespace: config.Namespace,
		scheme:    config.Scheme,
		store:     config.Store,
		table:     tb,
		loader:    ld,
	}
}

// LookupByID retrieves a symbol by ID from the table or loads it from the store if not found.
func (r *Runtime) Load(ctx context.Context, specs ...spec.Spec) ([]*symbol.Symbol, error) {
	for _, spec := range specs {
		spec.SetNamespace(r.namespace)
	}

	return r.loader.Load(ctx, specs...)
}

// Store adds a spec to the Runtime and returns the corresponding symbol.
func (r *Runtime) Store(ctx context.Context, specs ...spec.Spec) ([]*symbol.Symbol, error) {
	for _, spec := range specs {
		if spec.GetID() == uuid.Nil {
			spec.SetID(uuid.Must(uuid.NewV7()))
		}
		spec.SetNamespace(r.namespace)
	}

	exists := make(map[uuid.UUID]spec.Spec)
	if specs, err := r.store.Load(ctx, specs...); err != nil {
		return nil, err
	} else {
		for _, spec := range specs {
			exists[spec.GetID()] = spec
		}
	}

	for _, spec := range specs {
		if _, ok := exists[spec.GetID()]; ok {
			if _, err := r.store.Swap(ctx, spec); err != nil {
				return nil, err
			}
		} else {
			if _, err := r.store.Store(ctx, spec); err != nil {
				return nil, err
			}
		}
	}

	return r.loader.Load(ctx, specs...)
}

// Delete removes a spec from the Runtime and returns whether it was successfully deleted.
func (r *Runtime) Delete(ctx context.Context, specs ...spec.Spec) (int, error) {
	for _, spec := range specs {
		spec.SetNamespace(r.namespace)
	}

	specs, err := r.store.Load(ctx, specs...)
	if err != nil {
		return 0, err
	}

	count, err := r.store.Delete(ctx, specs...)
	if err != nil {
		return 0, err
	}

	for _, spec := range specs {
		if _, err := r.table.Free(spec.GetID()); err != nil {
			return 0, err
		}
	}
	return count, nil
}

// Listen starts the loader's watch process and reconciles symbols.
func (r *Runtime) Listen(ctx context.Context) error {
	spec := &spec.Meta{
		Namespace: r.namespace,
	}

	if err := r.loader.Watch(ctx, spec); err != nil {
		return err
	}
	if err := r.table.Clear(); err != nil {
		return err
	}
	if _, err := r.loader.Load(ctx, spec); err != nil {
		return err
	}
	return r.loader.Reconcile(ctx)
}

// Close shuts down the Runtime by closing the loader and clearing the symbol table.
func (r *Runtime) Close() error {
	if err := r.loader.Close(); err != nil {
		return err
	}
	return r.table.Clear()
}
