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
	Table  *Table         // Symbol table for storing loaded symbols
	Scheme *scheme.Scheme // Scheme for decoding and compiling specs
	Store  spec.Store     // Store to retrieve specs from
}

// Loader synchronizes with spec.Store to load spec.Spec into the Table.
type Loader struct {
	table  *Table
	scheme *scheme.Scheme
	store  spec.Store
	stream spec.Stream
	mu     sync.RWMutex
}

// NewLoader creates a new Loader instance with the provided configuration.
func NewLoader(config LoaderConfig) *Loader {
	return &Loader{
		table:  config.Table,
		scheme: config.Scheme,
		store:  config.Store,
	}
}

// Load loads a spec.Spec by ID and its linked specs into the symbol table.
func (l *Loader) Load(ctx context.Context, specs ...spec.Spec) ([]*Symbol, error) {
	var symbols []*Symbol
	nexts := specs
	for len(nexts) > 0 {
		curr := nexts
		nexts = nil

		specs, err := l.store.Load(ctx, curr...)
		if err != nil {
			return nil, err
		}

		for _, s := range specs {
			decode, err := l.scheme.Decode(s)
			if err != nil {
				return nil, err
			}

			sym, ok := l.table.Lookup(decode.GetID())
			if !ok || !reflect.DeepEqual(sym.Spec, decode) {
				n, err := l.scheme.Compile(decode)
				if err != nil {
					return nil, err
				}

				sym = &Symbol{Spec: decode, Node: n}
				if err := l.table.Insert(sym); err != nil {
					return nil, err
				}

				for _, locations := range sym.Ports() {
					for _, location := range locations {
						nexts = append(nexts, &spec.Meta{
							ID:        location.ID,
							Namespace: s.GetNamespace(),
							Name:      location.Name,
						})
					}
				}
			}

			symbols = append(symbols, sym)
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
				sym, ok := l.table.Lookup(spec.GetID())
				if ok {
					if _, err := l.table.Free(sym.ID()); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return symbols, nil
}

// Watch starts watching for changes to spec.Spec.
func (l *Loader) Watch(ctx context.Context, specs ...spec.Spec) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.stream != nil {
		return nil
	}

	s, err := l.store.Watch(ctx, specs...)
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

	unloaded := map[uuid.UUID]spec.Spec{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-stream.Next():
			if !ok {
				return nil
			}

			specs, err := l.store.Load(ctx, &spec.Meta{ID: event.ID})
			if err != nil {
				return err
			}

			for _, spec := range specs {
				unloaded[spec.GetID()] = spec
			}

			var examples []spec.Spec
			for _, example := range unloaded {
				examples = append(examples, &spec.Meta{ID: example.GetID()})
			}

			symbols, err := l.Load(ctx, examples...)
			if err != nil {
				return err
			}

			for _, sym := range symbols {
				delete(unloaded, sym.ID())
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
