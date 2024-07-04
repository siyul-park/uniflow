package loader

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Config contains the configuration settings for the Loader.
type Config struct {
	Namespace string        // Namespace associated with the Loader
	Table     *symbol.Table // Symbol table for storing loaded symbols
	Store     *store.Store  // Store to retrieve spec.Spec from
}

// Loader loads spec.Spec into the symbol.Table.
type Loader struct {
	namespace string
	table     *symbol.Table
	store     *store.Store
	stream    *store.Stream
	mu        sync.RWMutex
}

// New creates a new Loader instance with the given configuration.
func New(config Config) *Loader {
	return &Loader{
		namespace: config.Namespace,
		table:     config.Table,
		store:     config.Store,
	}
}

// LoadOne loads a single spec.Spec by ID, including linked specs, and adds them to the symbol table.
func (l *Loader) LoadOne(ctx context.Context, id uuid.UUID) (*symbol.Symbol, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	namespace := l.namespace
	nexts := []any{id}

	for len(nexts) > 0 {
		keys := nexts
		nexts = nil

		exists := map[any]bool{}
		var filter *store.Filter

		for _, key := range keys {
			exists[key] = false

			switch k := key.(type) {
			case uuid.UUID:
				filter = filter.Or(store.Where[uuid.UUID](spec.KeyID).EQ(k))
			case string:
				filter = filter.Or(store.Where[string](spec.KeyName).EQ(k))
			}
		}

		if namespace != "" {
			filter = filter.And(store.Where[string](spec.KeyNamespace).EQ(namespace))
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

			if sym, err := l.table.Insert(spec); err != nil {
				return nil, err
			} else if sym == nil {
				continue
			}

			for _, locations := range spec.GetLinks() {
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

// LoadAll loads all spec.Spec from the store and adds them to the symbol table.
func (l *Loader) LoadAll(ctx context.Context) ([]*symbol.Symbol, error) {
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
		filter = store.Where[string](spec.KeyNamespace).EQ(l.namespace)
	}

	specs, err := l.store.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}

	var symbols []*symbol.Symbol
	for _, spec := range specs {
		if sym, err := l.table.Insert(spec); err != nil {
			return nil, err
		} else if sym != nil {
			symbols = append(symbols, sym)
		} else if sym, ok := l.table.LookupByID(spec.GetID()); ok {
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
		filter = store.Where[string](spec.KeyNamespace).EQ(l.namespace)
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
