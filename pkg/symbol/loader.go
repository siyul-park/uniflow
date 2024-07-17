package symbol

import (
	"context"
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
)

// LoaderConfig holds configuration for the Loader.
type LoaderConfig struct {
	Namespace string       // Namespace for the Loader
	Table     *Table       // Symbol table for storing loaded symbols
	Store     *store.Store // Store to retrieve specs from
}

// Loader synchronizes with store.Store to load spec.Spec into the Table.
type Loader struct {
	namespace string
	table     *Table
	store     *store.Store
	stream    *store.Stream
	mu        sync.RWMutex
}

// NewLoader creates a new Loader instance with the provided configuration.
func NewLoader(config LoaderConfig) *Loader {
	return &Loader{
		namespace: config.Namespace,
		table:     config.Table,
		store:     config.Store,
	}
}

// LoadOne loads a spec.Spec by ID and its linked specs into the symbol table.
func (l *Loader) LoadOne(ctx context.Context, id uuid.UUID) (*Symbol, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	namespace := l.namespace
	nexts := []interface{}{id}

	for len(nexts) > 0 {
		keys := nexts
		nexts = nil

		exists := map[interface{}]bool{}
		var filter *store.Filter

		for _, key := range keys {
			exists[key] = false

			switch k := key.(type) {
			case uuid.UUID:
				filter = filter.Or(store.Where[uuid.UUID](spec.KeyID).Equal(k))
			case string:
				filter = filter.Or(store.Where[string](spec.KeyName).Equal(k))
			}
		}

		if namespace != "" {
			filter = filter.And(store.Where[string](spec.KeyNamespace).Equal(namespace))
		}

		specs, err := l.store.FindMany(ctx, filter, &database.FindOptions{Limit: lo.ToPtr(len(keys))})
		if err != nil {
			return nil, err
		}

		for _, spec := range specs {
			exists[spec.GetID()] = true
			if spec.GetName() != "" {
				exists[spec.GetName()] = true
			}

			if namespace == "" {
				namespace = spec.GetNamespace()
			}

			if sym, ok := l.table.LookupByID(spec.GetID()); ok && reflect.DeepEqual(sym.Spec, spec) {
				continue
			}

			sym := &Symbol{Spec: spec}
			if err := l.table.Insert(sym); err != nil {
				return nil, err
			}

			for _, locations := range sym.Links() {
				for _, location := range locations {
					if location.ID != (uuid.UUID{}) {
						nexts = append(nexts, location.ID)
					} else if location.Name != "" {
						nexts = append(nexts, location.Name)
					}
				}
			}
		}

		for key, exist := range exists {
			if !exist {
				id, ok := key.(uuid.UUID)
				if !ok {
					if name, ok := key.(string); ok {
						if sym, ok := l.table.LookupByName(namespace, name); ok {
							id = sym.ID()
						}
					}
				}

				if id != (uuid.UUID{}) {
					if _, err := l.table.Free(id); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	if sym, ok := l.table.LookupByID(id); !ok {
		return nil, nil
	} else {
		return sym, nil
	}
}

// LoadAll loads all spec.Spec from the store into the symbol table.
func (l *Loader) LoadAll(ctx context.Context) ([]*Symbol, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, id := range l.table.Keys() {
		if sym, ok := l.table.LookupByID(id); ok {
			if l.namespace == "" || sym.Namespace() == l.namespace {
				if _, err := l.table.Free(sym.ID()); err != nil {
					return nil, err
				}
			}
		}
	}

	var filter *store.Filter
	if l.namespace != "" {
		filter = store.Where[string](spec.KeyNamespace).Equal(l.namespace)
	}

	specs, err := l.store.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}

	var symbols []*Symbol
	for _, spec := range specs {
		sym := &Symbol{Spec: spec}
		if err := l.table.Insert(sym); err != nil {
			return nil, err
		} else {
			symbols = append(symbols, sym)
		}
	}

	return symbols, nil
}

// Watch starts watching for changes to spec.Spec.
func (l *Loader) Watch(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.stream != nil {
		return nil
	}

	var filter *store.Filter
	if l.namespace != "" {
		filter = store.Where[string](spec.KeyNamespace).Equal(l.namespace)
	}

	s, err := l.store.Watch(ctx, filter)
	if err != nil {
		return err
	}

	go func() {
		<-s.Done()

		l.mu.Lock()
		defer l.mu.Unlock()

		if l.stream == s {
			l.stream = nil
		}
	}()

	l.stream = s
	return nil
}

// Reconcile syncs changes to spec.Spec in the symbol table.
func (l *Loader) Reconcile(ctx context.Context) error {
	l.mu.RLock()
	stream := l.stream
	l.mu.RUnlock()

	if stream == nil {
		return nil
	}

	var nexts []uuid.UUID
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-stream.Next():
			if !ok {
				return nil
			}

			nexts = append(nexts, event.NodeID)

			for i := len(nexts) - 1; i >= 0; i-- {
				id := nexts[i]
				if _, err := l.LoadOne(ctx, id); err == nil {
					nexts = append(nexts[:i], nexts[i+1:]...)
				}
			}
		}
	}
}

// Close stops the loader and closes the associated stream.
func (l *Loader) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.stream == nil {
		return nil
	}

	if err := l.stream.Close(); err != nil {
		return err
	}
	l.stream = nil

	return nil
}
