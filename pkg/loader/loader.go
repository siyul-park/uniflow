package loader

import (
	"context"
	"reflect"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

type (
	// Config is a config for for the Loader.
	Config struct {
		Namespace string
		Table     *symbol.Table
		Scheme    *scheme.Scheme
		Storage   *storage.Storage
	}

	// Loader loads scheme.Spec into symbol.Table.
	Loader struct {
		namespace string
		scheme    *scheme.Scheme
		table     *symbol.Table
		storage   *storage.Storage
		mu        sync.RWMutex
	}
)

// New returns a new Loader.
func New(config Config) *Loader {
	namespace := config.Namespace
	table := config.Table
	scheme := config.Scheme
	storage := config.Storage

	return &Loader{
		namespace: namespace,
		scheme:    scheme,
		table:     table,
		storage:   storage,
	}
}

// LoadOne loads a single scheme.Spec from the storage.Storage
func (ld *Loader) LoadOne(ctx context.Context, id ulid.ULID) (node.Node, error) {
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
			if k, ok := key.(ulid.ULID); ok {
				exists[k] = false
				filter = filter.Or(storage.Where[ulid.ULID](scheme.KeyID).EQ(k))
			} else if k, ok := key.(string); ok {
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

			if sym, ok := ld.table.LookupByID(spec.GetID()); ok {
				if reflect.DeepEqual(sym.Spec, spec) {
					continue
				}
			}

			if n, err := ld.scheme.Decode(spec); err != nil {
				return nil, err
			} else if err := ld.table.Insert(&symbol.Symbol{Node: n, Spec: spec}); err != nil {
				return nil, err
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
			if exist {
				continue
			}

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

	if sym, ok := ld.table.LookupByID(id); !ok {
		return nil, nil
	} else {
		return sym.Node, nil
	}
}

// LoadAll loads all scheme.Spec from the storage.Storage
func (ld *Loader) LoadAll(ctx context.Context) ([]node.Node, error) {
	specs, err := ld.storage.FindMany(ctx, nil)
	if err != nil {
		return nil, err
	}

	var nodes []node.Node
	for _, spec := range specs {
		if sym, ok := ld.table.LookupByID(spec.GetID()); ok {
			if reflect.DeepEqual(sym.Spec, spec) {
				nodes = append(nodes, sym.Node)
				continue
			}
		}

		if n, err := ld.scheme.Decode(spec); err != nil {
			return nil, err
		} else if err := ld.table.Insert(&symbol.Symbol{Node: n, Spec: spec}); err != nil {
			return nil, err
		} else {
			nodes = append(nodes, n)
		}
	}

	return nodes, nil
}
