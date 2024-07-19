package symbol

import (
	"context"
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// LoaderConfig holds configuration for the Loader.
type LoaderConfig struct {
	Namespace string         // Namespace for the Loader
	Table     *Table         // Symbol table for storing loaded symbols
	Scheme    *scheme.Scheme // Scheme for decoding and compiling specs
	Store     *spec.Store    // Store to retrieve specs from
}

// Loader synchronizes with spec.Store to load spec.Spec into the Table.
type Loader struct {
	namespace string
	table     *Table
	scheme    *scheme.Scheme
	store     *spec.Store
	stream    *spec.Stream
	mu        sync.RWMutex
}

// NewLoader creates a new Loader instance with the provided configuration.
func NewLoader(config LoaderConfig) *Loader {
	return &Loader{
		namespace: config.Namespace,
		table:     config.Table,
		scheme:    config.Scheme,
		store:     config.Store,
	}
}

// LoadOne loads a spec.Spec by ID and its linked specs into the symbol table.
func (l *Loader) LoadOne(ctx context.Context, id uuid.UUID) (*Symbol, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	namespace := l.namespace
	nexts := []spec.Spec{&spec.Meta{ID: id, Namespace: namespace}}
	for len(nexts) > 0 {
		curr := nexts
		nexts = nil

		specs, err := l.store.Load(ctx, curr...)
		if err != nil {
			return nil, err
		}

		for _, s := range specs {
			if namespace == "" {
				namespace = s.GetNamespace()
			}

			decode, err := l.scheme.Decode(s)
			if err != nil {
				return nil, err
			}

			if sym, ok := l.table.LookupByID(decode.GetID()); ok && reflect.DeepEqual(sym.Spec, decode) {
				continue
			}

			n, err := l.scheme.Compile(decode)
			if err != nil {
				return nil, err
			}

			sym := &Symbol{Spec: decode, Node: n}
			if err := l.table.Insert(sym); err != nil {
				return nil, err
			}

			for _, locations := range sym.Links() {
				for _, location := range locations {
					nexts = append(nexts, &spec.Meta{
						ID:        location.ID,
						Namespace: namespace,
						Name:      location.Name,
					})
				}
			}
		}

		for _, spec := range curr {
			exists := false
			for _, s := range specs {
				if spec.GetID() == s.GetID() || (spec.GetNamespace() == s.GetNamespace() && spec.GetName() == s.GetName()) {
					exists = true
					break
				}
				if exists {
					break
				}
			}

			if !exists {
				sym, ok := l.table.LookupByID(spec.GetID())
				if !ok && spec.GetName() != "" {
					sym, ok = l.table.LookupByName(spec.GetNamespace(), spec.GetName())
				}
				if ok {
					if _, err := l.table.Free(sym.ID()); err != nil {
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

	specs, err := l.store.Load(ctx, &spec.Meta{
		Namespace: l.namespace,
	})
	if err != nil {
		return nil, err
	}

	var symbols []*Symbol
	for _, spec := range specs {
		spec, err := l.scheme.Decode(spec)
		if err != nil {
			return nil, err
		}

		if sym, ok := l.table.LookupByID(spec.GetID()); ok && reflect.DeepEqual(sym.Spec, spec) {
			continue
		}

		n, err := l.scheme.Compile(spec)
		if err != nil {
			return nil, err
		}

		sym := &Symbol{Spec: spec, Node: n}
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

	s, err := l.store.Watch(ctx, &spec.Meta{
		Namespace: l.namespace,
	})
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

			nexts = append(nexts, event.ID)

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
