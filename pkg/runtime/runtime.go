package runtime

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/value"
)

// Config defines configuration options for the Runtime.
type Config struct {
	Namespace   string            // Namespace defines the isolated execution environment for workflows.
	Environment map[string]string // Environment holds the variables for the loader.
	Hook        *hook.Hook        // Hook is a collection of hook functions for managing symbols.
	Scheme      *scheme.Scheme    // Scheme defines the scheme and behaviors for symbols.
	SpecStore   spec.Store        // SpecStore is responsible for persisting specifications.
	ValueStore  value.Store       // ValueStore is responsible for persisting values.
}

// Runtime represents an environment for executing Workflows.
type Runtime struct {
	namespace    string
	scheme       *scheme.Scheme
	specStore    spec.Store
	valueStore   value.Store
	specStream   spec.Stream
	valueStream  value.Stream
	symbolTable  *symbol.Table
	symbolLoader *symbol.Loader
	mu           sync.RWMutex
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
	if config.ValueStore == nil {
		config.ValueStore = value.NewStore()
	}

	config.Hook.AddLoadHook(symbol.LoadListenerHook(config.Hook))
	config.Hook.AddUnloadHook(symbol.UnloadListenerHook(config.Hook))

	symbolTable := symbol.NewTable(symbol.TableOption{
		LoadHooks:   []symbol.LoadHook{config.Hook},
		UnloadHooks: []symbol.UnloadHook{config.Hook},
	})
	symbolLoader := symbol.NewLoader(symbol.LoaderConfig{
		Environment: config.Environment,
		Table:       symbolTable,
		Scheme:      config.Scheme,
		SpecStore:   config.SpecStore,
		ValueStore:  config.ValueStore,
	})

	return &Runtime{
		namespace:    config.Namespace,
		scheme:       config.Scheme,
		specStore:    config.SpecStore,
		valueStore:   config.ValueStore,
		symbolTable:  symbolTable,
		symbolLoader: symbolLoader,
	}
}

// Load loads symbols from the spec store into the symbol table.
func (r *Runtime) Load(ctx context.Context) error {
	return r.symbolLoader.Load(ctx, &spec.Meta{Namespace: r.namespace})
}

// Watch sets up watchers for specification and value changes.
func (r *Runtime) Watch(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.specStream != nil {
		if err := r.specStream.Close(); err != nil {
			return err
		}
	}
	specStream, err := r.specStore.Watch(ctx, &spec.Meta{Namespace: r.namespace})
	if err != nil {
		return err
	}
	r.specStream = specStream

	if r.valueStream != nil {
		if err := r.valueStream.Close(); err != nil {
			return err
		}
	}
	valueStream, err := r.valueStore.Watch(ctx, &value.Value{Namespace: r.namespace})
	if err != nil {
		return err
	}
	r.valueStream = valueStream

	return nil
}

// Reconcile reconciles the state of symbols based on changes in specifications and values.
func (r *Runtime) Reconcile(ctx context.Context) error {
	r.mu.RLock()

	specStream := r.specStream
	valueStream := r.valueStream

	r.mu.RUnlock()

	if specStream == nil || valueStream == nil {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-specStream.Next():
			if !ok {
				return nil
			}

			specs, err := r.specStore.Load(ctx, &spec.Meta{ID: event.ID})
			if err != nil {
				return err
			}
			if len(specs) == 0 {
				if sb := r.symbolTable.Lookup(event.ID); sb != nil {
					specs = append(specs, sb.Spec)
				}
			}

			_ = r.symbolLoader.Load(ctx, specs...)
		case event, ok := <-valueStream.Next():
			if !ok {
				return nil
			}

			values, err := r.valueStore.Load(ctx, &value.Value{ID: event.ID})
			if err != nil {
				return err
			}
			if len(values) == 0 {
				values = append(values, &value.Value{ID: event.ID})
			}

			var specs []spec.Spec
			for _, id := range r.symbolTable.Keys() {
				if sb := r.symbolTable.Lookup(id); sb != nil {
					unstructured := &spec.Unstructured{}
					if err := spec.As(sb.Spec, unstructured); err != nil {
						return err
					} else if unstructured.IsBound(values...) {
						specs = append(specs, sb.Spec)
					}
				}
			}

			_ = r.symbolLoader.Load(ctx, specs...)
		}
	}
}

// Close shuts down the Runtime by closing streams and clearing the symbol table.
func (r *Runtime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.specStream != nil {
		if err := r.specStream.Close(); err != nil {
			return err
		}
		r.specStream = nil
	}
	if r.valueStream != nil {
		if err := r.valueStream.Close(); err != nil {
			return err
		}
		r.valueStream = nil
	}
	return r.symbolTable.Close()
}
