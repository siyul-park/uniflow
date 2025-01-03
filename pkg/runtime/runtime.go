package runtime

import (
	"context"
	"errors"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/value"
	"golang.org/x/exp/slices"
)

// Config defines configuration options for the Runtime.
type Config struct {
	Namespace   string            // Namespace defines the isolated execution environment for workflows.
	Environment map[string]string // Environment holds the variables for the loader.
	Hook        *hook.Hook        // Hook is a collection of hook functions for managing symbols.
	Scheme      *scheme.Scheme    // Scheme defines the scheme and behaviors for symbols.
	SpecStore   spec.Store        // SpecStore is responsible for persisting specifications.
	ValueStore  value.Store       // ValueStore is responsible for persisting values.
	ChartStore  chart.Store       // ChartStore is responsible for persisting charts.
}

// Runtime represents an environment for executing Workflows.
type Runtime struct {
	namespace    string
	scheme       *scheme.Scheme
	specStore    spec.Store
	valueStore   value.Store
	chartStore   chart.Store
	specStream   spec.Stream
	valueStream  value.Stream
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
	if config.ValueStore == nil {
		config.ValueStore = value.NewStore()
	}
	if config.ChartStore == nil {
		config.ChartStore = chart.NewStore()
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

	chartLinker := chart.NewLinker(config.Scheme)

	config.Hook.AddLinkHook(chartLinker)
	config.Hook.AddUnlinkHook(chartLinker)

	chartTable := chart.NewTable(chart.TableOption{
		LinkHooks:   []chart.LinkHook{config.Hook},
		UnlinkHooks: []chart.UnlinkHook{config.Hook},
	})
	chartLoader := chart.NewLoader(chart.LoaderConfig{
		Table:      chartTable,
		ChartStore: config.ChartStore,
		ValueStore: config.ValueStore,
	})

	for _, kind := range config.Scheme.Kinds() {
		_ = chartTable.Insert(&chart.Chart{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: config.Namespace,
			Name:      kind,
		})
	}

	return &Runtime{
		namespace:    config.Namespace,
		scheme:       config.Scheme,
		specStore:    config.SpecStore,
		valueStore:   config.ValueStore,
		chartStore:   config.ChartStore,
		symbolTable:  symbolTable,
		symbolLoader: symbolLoader,
		chartTable:   chartTable,
		chartLoader:  chartLoader,
	}
}

// Load loads symbols from the spec store into the symbol table.
func (r *Runtime) Load(ctx context.Context) error {
	var errs []error
	errs = append(errs, r.chartLoader.Load(ctx, &chart.Chart{Namespace: r.namespace}))
	errs = append(errs, r.symbolLoader.Load(ctx, &spec.Meta{Namespace: r.namespace}))
	return errors.Join(errs...)
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

// Reconcile reconciles the state of symbols based on changes in specifications and values.
func (r *Runtime) Reconcile(ctx context.Context) error {
	r.mu.RLock()

	specStream := r.specStream
	valueStream := r.valueStream
	chartStream := r.chartStream

	r.mu.RUnlock()

	if specStream == nil || valueStream == nil || chartStream == nil {
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

			var specs []spec.Spec
			for _, id := range r.symbolTable.Keys() {
				sb := r.symbolTable.Lookup(id)
				if sb != nil && slices.Contains(kinds, sb.Kind()) {
					specs = append(specs, sb.Spec)
				}
			}

			for _, sp := range specs {
				_, _ = r.symbolTable.Free(sp.GetID())
			}

			_ = r.chartLoader.Load(ctx, &chart.Chart{ID: event.ID})
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
