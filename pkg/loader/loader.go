package loader

import (
	"context"
	"reflect"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Config represents the configuration for the Loader.
type Config struct {
	Namespace string           // Namespace is the namespace used by the Loader.
	Table     *symbol.Table    // Table is the symbol table for managing symbols.
	Storage   *storage.Storage // Storage is the storage used by the Loader.
}

// Loader loads scheme.Spec into the symbol.Table.
type Loader struct {
	namespace string
	table     *symbol.Table
	storage   *storage.Storage
	mu        sync.RWMutex
}

// New returns a new Loader.
func New(config Config) *Loader {
	namespace := config.Namespace
	table := config.Table
	storage := config.Storage

	return &Loader{
		namespace: namespace,
		table:     table,
		storage:   storage,
	}
}

// LoadOne loads a single scheme.Spec from the storage.Storage.
// It processes the specified ID and recursively loads linked scheme.Spec.
// If the loader is associated with a namespace, it uses that namespace.
// The loaded nodes are added to the symbol table for future reference.
func (ld *Loader) LoadOne(ctx context.Context, id ulid.ULID) (*symbol.Symbol, error) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	namespace := ld.namespace
	queue := []any{id}

	for len(queue) > 0 {
		prev := queue
		queue = nil
		exists := map[any]bool{}
		var filter *storage.Filter

		for _, key := range prev {
			switch k := key.(type) {
			case ulid.ULID:
				exists[k] = false
				filter = filter.Or(storage.Where[ulid.ULID](scheme.KeyID).EQ(k))
			case string:
				exists[k] = false
				filter = filter.Or(storage.Where[string](scheme.KeyName).EQ(k))
			}
		}

		if namespace != "" {
			filter = filter.And(storage.Where[string](scheme.KeyNamespace).EQ(namespace))
		}

		specs, err := ld.storage.FindMany(ctx, filter, &database.FindOptions{Limit: lo.ToPtr(len(prev))})
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

			if sym, err := ld.table.Insert(spec); err != nil {
				return nil, err
			} else if sym.Spec() != spec {
				continue
			}

			for _, locations := range spec.GetLinks() {
				for _, location := range locations {
					if location.ID != (ulid.ULID{}) {
						queue = append(queue, location.ID)
					} else if location.Name != "" {
						queue = append(queue, location.Name)
					}
				}
			}
		}

		for key, exist := range exists {
			if !exist {
				id, ok := key.(ulid.ULID)
				if !ok {
					if name, ok := key.(string); ok {
						if sym, ok := ld.table.LookupByName(namespace, name); ok {
							id = sym.ID()
						}
					}
				}

				if id != (ulid.ULID{}) {
					if _, err := ld.table.Free(id); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	if sym, ok := ld.table.LookupByID(id); !ok {
		return nil, nil
	} else {
		return sym, nil
	}
}

// LoadAll loads all scheme.Spec from the storage.Storage.
// It loads all available scheme.Spec and adds them to the symbol table for future reference.
// If the loader is associated with a namespace, it filters the loading based on that namespace.
func (ld *Loader) LoadAll(ctx context.Context) ([]*symbol.Symbol, error) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	var filter *storage.Filter

	if ld.namespace != "" {
		filter = filter.And(storage.Where[string](scheme.KeyNamespace).EQ(ld.namespace))
	}

	specs, err := ld.storage.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}

	var symbols []*symbol.Symbol
	for _, spec := range specs {
		if sym, ok := ld.table.LookupByID(spec.GetID()); ok && reflect.DeepEqual(sym.Spec, spec) {
			symbols = append(symbols, sym)
			continue
		}

		if sym, err := ld.table.Insert(spec); err != nil {
			return nil, err
		} else {
			symbols = append(symbols, sym)
		}
	}

	return symbols, nil
}
