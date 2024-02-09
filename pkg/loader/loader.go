package loader

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Config contains the configuration settings for the Loader.
type Config struct {
	Namespace string           // Namespace associated with the Loader
	Table     *symbol.Table    // Symbol table for storing loaded symbols
	Storage   *storage.Storage // Storage to retrieve scheme.Spec from
}

// Loader is responsible for loading scheme.Spec into the symbol.Table.
type Loader struct {
	namespace string
	table     *symbol.Table
	storage   *storage.Storage
	mu        sync.RWMutex
}

// New creates a new Loader instance with the given configuration.
func New(config Config) *Loader {
	return &Loader{
		namespace: config.Namespace,
		table:     config.Table,
		storage:   config.Storage,
	}
}

// LoadOne loads a single scheme.Spec from storage.Storage.
// It recursively loads linked scheme.Spec based on the specified ID.
// If the Loader is associated with a namespace, it uses that namespace.
// Loaded symbols are added to the symbol table for future reference.
func (ld *Loader) LoadOne(ctx context.Context, id uuid.UUID) (*symbol.Symbol, error) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	namespace := ld.namespace
	next := []any{id}

	for len(next) > 0 {
		cur := next
		next = nil

		exists := map[any]bool{}
		var filter *storage.Filter

		for _, key := range cur {
			exists[key] = false

			switch k := key.(type) {
			case uuid.UUID:
				filter = filter.Or(storage.Where[uuid.UUID](scheme.KeyID).EQ(k))
			case string:
				filter = filter.Or(storage.Where[string](scheme.KeyName).EQ(k))
			}
		}

		if namespace != "" {
			filter = filter.And(storage.Where[string](scheme.KeyNamespace).EQ(namespace))
		}

		specs, err := ld.storage.FindMany(ctx, filter, &database.FindOptions{Limit: lo.ToPtr(len(cur))})
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
			} else if sym == nil {
				continue
			}

			for _, locations := range spec.GetLinks() {
				for _, location := range locations {
					if location.ID != (uuid.UUID{}) {
						next = append(next, location.ID)
					} else if location.Name != "" {
						next = append(next, location.Name)
					}
				}
			}
		}

		for key, exist := range exists {
			if !exist {
				id, ok := key.(uuid.UUID)
				if !ok {
					if name, ok := key.(string); ok {
						if sym, ok := ld.table.LookupByName(namespace, name); ok {
							id = sym.ID()
						}
					}
				}

				if id != (uuid.UUID{}) {
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
// It adds the retrieved scheme.Spec to the symbol table for future reference.
// If the loader is associated with a namespace, it filters the loading based on that namespace.
func (ld *Loader) LoadAll(ctx context.Context) ([]*symbol.Symbol, error) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	for _, id := range ld.table.Keys() {
		if sym, ok := ld.table.LookupByID(id); ok {
			if ld.namespace == "" || sym.Namespace() == ld.namespace {
				if _, err := ld.table.Free(sym.ID()); err != nil {
					return nil, err
				}
			}
		}
	}

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
		if sym, err := ld.table.Insert(spec); err != nil {
			return nil, err
		} else if sym != nil {
			symbols = append(symbols, sym)
		} else if sym, ok := ld.table.LookupByID(spec.GetID()); ok {
			symbols = append(symbols, sym)
		}
	}

	return symbols, nil
}
