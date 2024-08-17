package runtime

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/debug"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Config defines configuration options for the Runtime.
type Config struct {
	Namespace   string          // Namespace defines the isolated execution environment for workflows.
	Hook        *hook.Hook      // Hook is a collection of hook functions for managing symbols.
	Scheme      *scheme.Scheme  // Scheme defines the scheme and behaviors for symbols.
	SpecStore   spec.Store      // SpecStore is responsible for persisting symbols.
	SecretStore secret.Store    // SpecStore is responsible for persisting symbols.
	Debugger    *debug.Debugger // Debugger provides debugging capabilities.
}

// Runtime represents an environment for executing Workflows.
type Runtime struct {
	namespace   string
	scheme      *scheme.Scheme
	specStore   spec.Store
	secretStore secret.Store
	table       *symbol.Table
	loader      *symbol.Loader
}

// New creates a new Runtime instance with the specified configuration.
func New(config Config) *Runtime {
	if config.Namespace == "" {
		config.Namespace = resource.DefaultNamespace
	}
	if config.Hook == nil {
		config.Hook = hook.New()
	}
	if config.Scheme == nil {
		config.Scheme = scheme.New()
	}
	if config.SpecStore == nil {
		config.SpecStore = spec.NewStore()
	}
	if config.SecretStore == nil {
		config.SecretStore = secret.NewStore()
	}

	var loadHooks []symbol.LoadHook
	var unloadHooks []symbol.UnloadHook
	if config.Debugger != nil {
		loadHooks = append(loadHooks, config.Debugger)
		unloadHooks = append(unloadHooks, config.Debugger)
	}
	loadHooks = append(loadHooks, config.Hook)
	unloadHooks = append(unloadHooks, config.Hook)

	tb := symbol.NewTable(symbol.TableOptions{
		LoadHooks:   loadHooks,
		UnloadHooks: unloadHooks,
	})

	ld := symbol.NewLoader(symbol.LoaderConfig{
		Scheme:      config.Scheme,
		SpecStore:   config.SpecStore,
		SecretStore: config.SecretStore,
		Table:       tb,
	})

	return &Runtime{
		namespace:   config.Namespace,
		scheme:      config.Scheme,
		specStore:   config.SpecStore,
		secretStore: config.SecretStore,
		table:       tb,
		loader:      ld,
	}
}

// LookupByID retrieves a symbol by ID from the table or loads it from the store if not found.
func (r *Runtime) Load(ctx context.Context, specs ...spec.Spec) ([]*symbol.Symbol, error) {
	if len(specs) == 0 {
		specs = append(specs, &spec.Meta{
			Namespace: r.namespace,
		})
	}

	for _, spc := range specs {
		if spc.GetNamespace() != r.namespace {
			spc.SetNamespace(r.namespace)
		}
	}

	return r.loader.Load(ctx, specs...)
}

// Store adds a spec to the Runtime and returns the corresponding symbol.
func (r *Runtime) Store(ctx context.Context, specs ...spec.Spec) ([]*symbol.Symbol, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	for _, spc := range specs {
		if spc.GetID() == uuid.Nil {
			spc.SetID(uuid.Must(uuid.NewV7()))
		}
		if spc.GetNamespace() != r.namespace {
			spc.SetNamespace(r.namespace)
		}
	}

	exists := make(map[uuid.UUID]spec.Spec)
	if specs, err := r.specStore.Load(ctx, specs...); err != nil {
		return nil, err
	} else {
		for _, spc := range specs {
			exists[spc.GetID()] = spc
		}
	}

	for _, spc := range specs {
		if _, ok := exists[spc.GetID()]; ok {
			if _, err := r.specStore.Swap(ctx, spc); err != nil {
				return nil, err
			}
		} else {
			if _, err := r.specStore.Store(ctx, spc); err != nil {
				return nil, err
			}
		}
	}

	return r.loader.Load(ctx, specs...)
}

// Delete removes a spec from the Runtime and returns whether it was successfully deleted.
func (r *Runtime) Delete(ctx context.Context, specs ...spec.Spec) (int, error) {
	if len(specs) == 0 {
		specs = append(specs, &spec.Meta{
			Namespace: r.namespace,
		})
	}

	for _, spc := range specs {
		if spc.GetNamespace() != r.namespace {
			spc.SetNamespace(r.namespace)
		}
	}

	specs, err := r.specStore.Load(ctx, specs...)
	if err != nil {
		return 0, err
	}

	count, err := r.specStore.Delete(ctx, specs...)
	if err != nil {
		return 0, err
	}

	for _, spc := range specs {
		if _, err := r.table.Free(spc.GetID()); err != nil {
			return 0, err
		}
	}
	return count, nil
}

// Listen starts the loader's watch process and reconciles symbols.
func (r *Runtime) Listen(ctx context.Context) error {
	spc := &spec.Meta{Namespace: r.namespace}
	if err := r.loader.Watch(ctx, spc); err != nil {
		return err
	}
	if _, err := r.loader.Load(ctx, spc); err != nil {
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
