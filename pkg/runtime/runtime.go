package runtime

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"golang.org/x/exp/slices"
)

// Config defines configuration options for the Runtime.
type Config struct {
	Namespace   string            // Namespace defines the isolated execution environment for workflows.
	Environment map[string]string // Environment holds the variables for the loader.
	Hook        *hook.Hook        // Hook is a collection of hook functions for managing symbols.
	Scheme      *scheme.Scheme    // Scheme defines the scheme and behaviors for symbols.
	SpecStore   spec.Store        // SpecStore is responsible for persisting specifications.
	SecretStore secret.Store      // SecretStore is responsible for persisting secrets.
	ChartStore  chart.Store       // ChartStore is responsible for persisting charts.
}

// Runtime represents an environment for executing Workflows.
type Runtime struct {
	namespace    string
	scheme       *scheme.Scheme
	specStore    spec.Store
	secretStore  secret.Store
	chartStore   chart.Store
	specStream   spec.Stream
	secretStream secret.Stream
	chartStream  chart.Stream
	symbolTable  *symbol.Table
	symbolLoader *symbol.Loader
	chartTable   *chart.Table
	chartLoader  *chart.Loader
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
	if config.ChartStore == nil {
		config.ChartStore = chart.NewStore()
	}

	symbolTable := symbol.NewTable(symbol.TableOption{
		LoadHooks:   []symbol.LoadHook{config.Hook},
		UnloadHooks: []symbol.UnloadHook{config.Hook},
	})
	symbolLoader := symbol.NewLoader(symbol.LoaderConfig{
		Environment: config.Environment,
		Table:       symbolTable,
		Scheme:      config.Scheme,
		SpecStore:   config.SpecStore,
		SecretStore: config.SecretStore,
	})

	chartLinker := chart.NewLinker(chart.LinkerConfig{
		LoadHooks:   []symbol.LoadHook{config.Hook},
		UnloadHooks: []symbol.UnloadHook{config.Hook},
		Scheme:      config.Scheme,
	})
	chartTable := chart.NewTable(chart.TableOption{
		LinkHooks:   []chart.LinkHook{chartLinker, config.Hook},
		UnlinkHooks: []chart.UnlinkHook{chartLinker, config.Hook},
	})
	chartLoader := chart.NewLoader(chart.LoaderConfig{
		Table:       chartTable,
		ChartStore:  config.ChartStore,
		SecretStore: config.SecretStore,
	})

	for _, kind := range config.Scheme.Kinds() {
		chartTable.Insert(&chart.Chart{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: config.Namespace,
			Name:      kind,
		})
	}

	return &Runtime{
		namespace:    config.Namespace,
		scheme:       config.Scheme,
		specStore:    config.SpecStore,
		secretStore:  config.SecretStore,
		chartStore:   config.ChartStore,
		symbolTable:  symbolTable,
		symbolLoader: symbolLoader,
		chartTable:   chartTable,
		chartLoader:  chartLoader,
	}
}

// Load loads symbols from the spec store into the symbol table.
func (r *Runtime) Load(ctx context.Context) error {
	if err := r.chartLoader.Load(ctx, &chart.Chart{Namespace: r.namespace}); err != nil {
		return err
	}
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

	if r.chartStream != nil {
		if err := r.chartStream.Close(); err != nil {
			return err
		}
	}
	chartStream, err := r.chartStore.Watch(ctx, &chart.Chart{Namespace: r.namespace})
	if err != nil {
		return err
	}
	r.chartStream = chartStream

	return nil
}

// Reconcile reconciles the state of symbols based on changes in specifications and secrets.
func (r *Runtime) Reconcile(ctx context.Context) error {
	r.mu.RLock()

	specStream := r.specStream
	secretStream := r.secretStream
	chartStream := r.chartStream

	r.mu.RUnlock()

	if specStream == nil || secretStream == nil || chartStream == nil {
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
				if sb := r.symbolTable.Lookup(event.ID); sb != nil {
					specs = append(specs, sb.Spec)
				}
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
				sb := r.symbolTable.Lookup(id)
				if sb != nil && spec.IsBound(sb.Spec, secrets...) {
					bounded[sb.ID()] = sb.Spec
				}
			}
			for _, sp := range unloaded {
				if spec.IsBound(sp, secrets...) {
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
		case event, ok := <-chartStream.Next():
			if !ok {
				return nil
			}

			charts := r.chartTable.Links(event.ID)
			if len(charts) == 0 {
				var err error
				charts, err = r.chartStore.Load(ctx, &chart.Chart{ID: event.ID})
				if err != nil {
					return err
				}
			}

			kinds := make([]string, 0, len(charts))
			for _, chrt := range charts {
				kinds = append(kinds, chrt.GetName())
			}

			bounded := make(map[uuid.UUID]spec.Spec)
			for _, id := range r.symbolTable.Keys() {
				sb := r.symbolTable.Lookup(id)
				if sb != nil && slices.Contains(kinds, sb.Kind()) {
					bounded[sb.ID()] = sb.Spec
				}
			}
			for _, sp := range unloaded {
				if slices.Contains(kinds, sp.GetKind()) {
					bounded[sp.GetID()] = sp
				}
			}

			for _, sp := range bounded {
				r.symbolTable.Free(sp.GetID())
			}

			r.chartLoader.Load(ctx, &chart.Chart{ID: event.ID})

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
	if r.chartStream != nil {
		if err := r.chartStream.Close(); err != nil {
			return err
		}
		r.chartStream = nil
	}

	if err := r.chartTable.Close(); err != nil {
		return err
	}
	return r.symbolTable.Close()
}
