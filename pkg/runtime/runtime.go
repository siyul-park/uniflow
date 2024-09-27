package runtime

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Config defines configuration options for the Runtime.
type Config struct {
	Namespace   string         // Namespace defines the isolated execution environment for workflows.
	Hook        *hook.Hook     // Hook is a collection of hook functions for managing symbols.
	Scheme      *scheme.Scheme // Scheme defines the scheme and behaviors for symbols.
	SpecStore   spec.Store     // SpecStore is responsible for persisting specifications.
	SecretStore secret.Store   // SecretStore is responsible for persisting secrets.
}

// Runtime represents an environment for executing Workflows.
type Runtime struct {
	namespace    string
	scheme       *scheme.Scheme
	specStore    spec.Store
	secretStore  secret.Store
	specStream   spec.Stream
	secretStream secret.Stream
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
	if config.SecretStore == nil {
		config.SecretStore = secret.NewStore()
	}

	symbolTable := symbol.NewTable(symbol.TableOptions{
		LoadHooks:   []symbol.LoadHook{config.Hook},
		UnloadHooks: []symbol.UnloadHook{config.Hook},
	})
	symbolLoader := symbol.NewLoader(symbol.LoaderConfig{
		Scheme:      config.Scheme,
		SpecStore:   config.SpecStore,
		SecretStore: config.SecretStore,
		Table:       symbolTable,
	})

	return &Runtime{
		namespace:    config.Namespace,
		scheme:       config.Scheme,
		specStore:    config.SpecStore,
		secretStore:  config.SecretStore,
		symbolTable:  symbolTable,
		symbolLoader: symbolLoader,
	}
}

// Load loads symbols from the spec store into the symbol table.
func (r *Runtime) Load(ctx context.Context) error {
	return r.symbolLoader.Load(ctx, &spec.Meta{Namespace: r.namespace})
}

// Watch sets up watchers for specification and secret changes.
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

	if r.secretStream != nil {
		if err := r.secretStream.Close(); err != nil {
			return err
		}
	}
	secretStream, err := r.secretStore.Watch(ctx, &secret.Secret{Namespace: r.namespace})
	if err != nil {
		return err
	}
	r.secretStream = secretStream

	return nil
}

// Reconcile reconciles the state of symbols based on changes in specifications and secrets.
func (r *Runtime) Reconcile(ctx context.Context) error {
	r.mu.RLock()

	specStream := r.specStream
	secretStream := r.secretStream

	r.mu.RUnlock()

	if specStream == nil || secretStream == nil {
		return nil
	}

	unloaded := make(map[uuid.UUID]spec.Spec)

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
				specs = append(specs, &spec.Meta{ID: event.ID})
			}

			for _, sp := range specs {
				if err := r.symbolLoader.Load(ctx, sp); err != nil {
					unloaded[sp.GetID()] = sp
				} else {
					delete(unloaded, sp.GetID())
				}
			}
		case event, ok := <-secretStream.Next():
			if !ok {
				return nil
			}

			secrets, err := r.secretStore.Load(ctx, &secret.Secret{ID: event.ID})
			if err != nil {
				return err
			}
			if len(secrets) == 0 {
				secrets = append(secrets, &secret.Secret{ID: event.ID})
			}

			bounded := make(map[uuid.UUID]spec.Spec)
			for _, id := range r.symbolTable.Keys() {
				sb, ok := r.symbolTable.Lookup(id)
				if ok && r.scheme.IsBound(sb.Spec, secrets...) {
					bounded[sb.Spec.GetID()] = sb.Spec
				}
			}
			for _, sp := range unloaded {
				if r.scheme.IsBound(sp, secrets...) {
					bounded[sp.GetID()] = sp
				}
			}

			for _, sp := range bounded {
				if err := r.symbolLoader.Load(ctx, sp); err != nil {
					unloaded[sp.GetID()] = sp
				} else {
					delete(unloaded, sp.GetID())
				}
			}
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
	if r.secretStream != nil {
		if err := r.secretStream.Close(); err != nil {
			return err
		}
		r.secretStream = nil
	}

	return r.symbolTable.Close()
}
