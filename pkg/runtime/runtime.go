package runtime

import (
	"context"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/types"
	"reflect"
	"slices"
	"sync"

	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/value"
)

// Config defines configuration options for the Runtime.
type Config struct {
	Namespace   string            // Namespace defines the isolated execution environment for workflows.
	Environment map[string]string // Environment holds the variables for the loader.
	Hook        *hook.Hook        // Hook is a collection of hook functions for managing symbols.
	Scheme      *scheme.Scheme    // Scheme defines the scheme and behaviors for symbols.
	SpecStore   store.Store       // SpecStore is responsible for persisting specifications.
	ValueStore  store.Store       // ValueStore is responsible for persisting values.
}

// Runtime represents an environment for executing Workflows.
type Runtime struct {
	namespace   string
	environment map[string]string
	scheme      *scheme.Scheme
	specStore   store.Store
	valueStore  store.Store
	specStream  store.Stream
	valueStream store.Stream
	symbolTable *symbol.Table
	mu          sync.RWMutex
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
		config.SpecStore = store.New()
	}
	if config.ValueStore == nil {
		config.ValueStore = store.New()
	}

	config.Hook.AddLoadHook(symbol.LoadListenerHook(config.Hook))
	config.Hook.AddUnloadHook(symbol.UnloadListenerHook(config.Hook))

	symbolTable := symbol.NewTable(symbol.TableOption{
		LoadHooks:   []symbol.LoadHook{config.Hook},
		UnloadHooks: []symbol.UnloadHook{config.Hook},
	})

	return &Runtime{
		namespace:   config.Namespace,
		environment: config.Environment,
		scheme:      config.Scheme,
		specStore:   config.SpecStore,
		valueStore:  config.ValueStore,
		symbolTable: symbolTable,
	}
}

// Load loads symbols from the spec store into the symbol table.
func (r *Runtime) Load(ctx context.Context, filter types.Map) error {
	filter = store.And(filter, store.Where(spec.KeyNamespace).Equal(types.NewString(r.namespace)))

	docs, err := r.specStore.Find(ctx, filter)
	if err != nil {
		return err
	}

	specs := make([]*spec.Unstructured, 0, len(docs))
	for _, doc := range docs {
		unstructured := &spec.Unstructured{}
		if err := types.Unmarshal(doc, unstructured); err != nil {
			return err
		}
		specs = append(specs, unstructured)
	}

	var filters []types.Map
	for _, sp := range specs {
		for _, val := range sp.GetEnv() {
			if val.ID != uuid.Nil {
				id, err := types.Marshal(val.ID)
				if err != nil {
					return err
				}
				filters = append(filters, store.Where(value.KeyID).Equal(id))
			}
			if val.Name != "" {
				filters = append(filters, store.And(store.Where(value.KeyNamespace).Equal(types.NewString(sp.GetNamespace())), store.Where(value.KeyName).Equal(types.NewString(val.Name))))
			}
		}
	}

	docs, err = r.valueStore.Find(ctx, store.Or(filters...))
	if err != nil {
		return err
	}

	values := make([]*value.Value, 0, len(docs))
	for _, doc := range docs {
		val := &value.Value{}
		if err := types.Unmarshal(doc, val); err != nil {
			return err
		}
		values = append(values, val)
	}

	if len(r.environment) > 0 {
		values = append(values, &value.Value{Data: r.environment})
	}

	var symbols []*symbol.Symbol
	var errs []error
	for _, unstructured := range specs {
		sp := spec.Spec(unstructured)
		if err := unstructured.Bind(values...); err != nil {
			errs = append(errs, err)
		} else if err := unstructured.Build(); err != nil {
			errs = append(errs, err)
		} else if decode, err := r.scheme.Decode(unstructured); err != nil {
			errs = append(errs, err)
		} else {
			sp = decode
		}

		sb := r.symbolTable.Lookup(sp.GetID())
		if sb == nil || !reflect.DeepEqual(sb.Spec, sp) {
			var n node.Node
			if sp != unstructured {
				if n, err = r.scheme.Compile(sp); err != nil {
					errs = append(errs, err)
				}
			}

			sb = &symbol.Symbol{Spec: unstructured, Node: n}
			if err := r.symbolTable.Insert(sb); err != nil {
				errs = append(errs, err)
			}
		}

		symbols = append(symbols, sb)
	}

	for _, id := range r.symbolTable.Keys() {
		sb := r.symbolTable.Lookup(id)
		if sb == nil {
			continue
		}

		doc, err := types.Cast[types.Map](types.Marshal(sb.Spec))
		if err != nil {
			errs = append(errs, err)
			continue
		}

		local := store.New()

		if err := local.Insert(ctx, []types.Map{doc}); err != nil {
			errs = append(errs, err)
			continue
		}

		docs, err := local.Find(ctx, filter)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if !slices.Contains(docs, doc) {
			continue
		}

		ok := false
		for _, s := range symbols {
			if s.ID() == id {
				ok = true
				break
			}
		}
		if !ok {
			if _, err := r.symbolTable.Free(id); err != nil {
				errs = append(errs, err)
			}
		}
	}
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
	specStream, err := r.specStore.Watch(ctx, store.Where(spec.KeyNamespace).Equal(types.NewString(r.namespace)))
	if err != nil {
		return err
	}
	r.specStream = specStream

	if r.valueStream != nil {
		if err := r.valueStream.Close(); err != nil {
			return err
		}
	}
	valueStream, err := r.valueStore.Watch(ctx, store.Where(value.KeyNamespace).Equal(types.NewString(r.namespace)))
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

			_ = r.Load(ctx, store.Where(spec.KeyID).Equal(event.Get(types.NewString(spec.KeyID))))
		case event, ok := <-valueStream.Next():
			if !ok {
				return nil
			}

			docs, err := r.valueStore.Find(ctx, store.Where(value.KeyID).Equal(event.Get(types.NewString(value.KeyID))))
			if err != nil {
				return err
			}

			values := make([]*value.Value, 0, len(docs))
			for _, doc := range docs {
				val := &value.Value{}
				if err := types.Unmarshal(doc, val); err != nil {
					return err
				}
				values = append(values, val)
			}

			var filters []types.Map
			for _, id := range r.symbolTable.Keys() {
				if sb := r.symbolTable.Lookup(id); sb != nil {
					unstructured := &spec.Unstructured{}
					if err := spec.As(sb.Spec, unstructured); err != nil {
						return err
					} else if unstructured.IsBound(values...) {
						id, err := types.Marshal(sb.ID())
						if err != nil {
							return err
						}
						filters = append(filters, store.Where(spec.KeyID).Equal(id))
					}
				}
			}

			_ = r.Load(ctx, store.Or(filters...))
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
