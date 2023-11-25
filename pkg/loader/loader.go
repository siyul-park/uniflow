package loader

import (
	"context"
	"reflect"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

type (
	// Config is a config for for the Loader.
	Config struct {
		Table   *symbol.Table
		Scheme  *scheme.Scheme
		Storage *storage.Storage
	}

	// Loader loads scheme.Spec into symbol.Table.
	Loader struct {
		scheme     *scheme.Scheme
		table      *symbol.Table
		remote     *storage.Storage
		local      *storage.Storage
		referenced map[ulid.ULID]links
		undefined  map[ulid.ULID]links
		mu         sync.RWMutex
	}

	links map[string][]scheme.PortLocation
)

// New returns a new Loader.
func New(ctx context.Context, config Config) (*Loader, error) {
	table := config.Table
	scheme := config.Scheme
	remote := config.Storage

	local, err := storage.New(ctx, storage.Config{
		Scheme:   scheme,
		Database: memdb.New(""),
	})
	if err != nil {
		return nil, err
	}

	return &Loader{
		scheme:     scheme,
		table:      table,
		remote:     remote,
		local:      local,
		referenced: make(map[ulid.ULID]links),
		undefined:  make(map[ulid.ULID]links),
	}, nil
}

// LoadOne loads a single scheme.Spec from the storage.Storage
func (ld *Loader) LoadOne(ctx context.Context, filter *storage.Filter) (node.Node, error) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	return ld.loadOne(ctx, filter)
}

// LoadMany loads multiple scheme.Spec from the storage.Storage
func (ld *Loader) LoadMany(ctx context.Context, filter *storage.Filter) ([]node.Node, error) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	return ld.loadMany(ctx, filter)
}

// UnloadOne unloads a single scheme.Spec from the storage.Storage
func (ld *Loader) UnloadOne(ctx context.Context, filter *storage.Filter) (bool, error) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	return ld.unloadOne(ctx, filter)
}

// UnloadMany unloads multiple scheme.Spec from the storage.Storage
func (ld *Loader) UnloadMany(ctx context.Context, filter *storage.Filter) (int, error) {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	return ld.unloadMany(ctx, filter)
}

func (ld *Loader) loadOne(ctx context.Context, filter *storage.Filter) (node.Node, error) {
	remote, err := ld.remote.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	local, err := ld.local.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	if remote != nil {
		if local != nil {
			if reflect.DeepEqual(remote, local) {
				if n, ok := ld.table.Lookup(remote.GetID()); ok {
					return n, nil
				}
			}
		}
	} else {
		if local != nil {
			_, err := ld.unloadOne(ctx, storage.Where[ulid.ULID](scheme.KeyID).EQ(local.GetID()))
			return nil, err
		}
		return nil, nil
	}

	if n, err := ld.scheme.Decode(remote); err != nil {
		return nil, err
	} else {
		err := ld.table.Insert(n, remote)
		if err != nil {
			return nil, err
		}

		if local == nil {
			if _, err := ld.local.InsertOne(ctx, remote); err != nil {
				return nil, err
			}
		} else {
			if _, err := ld.local.UpdateOne(ctx, remote); err != nil {
				return nil, err
			}
		}

		return n, nil
	}
}

func (ld *Loader) loadMany(ctx context.Context, filter *storage.Filter) ([]node.Node, error) {
	remotes, err := ld.remote.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	locals, err := ld.local.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}

	idToLocal := map[ulid.ULID]scheme.Spec{}
	idToRemote := map[ulid.ULID]scheme.Spec{}
	for _, spec := range locals {
		idToLocal[spec.GetID()] = spec
	}
	for _, spec := range remotes {
		idToRemote[spec.GetID()] = spec
	}

	var removeIds []ulid.ULID
	for id := range idToLocal {
		if _, ok := idToRemote[id]; !ok {
			removeIds = append(removeIds, id)
		}
	}
	if len(removeIds) > 0 {
		if _, err := ld.unloadMany(ctx, storage.Where[ulid.ULID](scheme.KeyID).IN(removeIds...)); err != nil {
			return nil, err
		}
	}

	var nodes []node.Node
	for id, remote := range idToRemote {
		local := idToLocal[id]
		if local != nil {
			if reflect.DeepEqual(remote, local) {
				if n, ok := ld.table.Lookup(id); ok {
					nodes = append(nodes, n)
					continue
				}
			}
		}

		if n, err := ld.scheme.Decode(remote); err != nil {
			return nil, err
		} else {
			if err := ld.table.Insert(n, remote); err != nil {
				return nil, err
			} else {
				nodes = append(nodes, n)
			}
			if local == nil {
				if _, err := ld.local.InsertOne(ctx, remote); err != nil {
					return nil, err
				}
			} else {
				if _, err := ld.local.UpdateOne(ctx, remote); err != nil {
					return nil, err
				}
			}
		}
	}

	return nodes, nil
}

func (ld *Loader) unloadOne(ctx context.Context, filter *storage.Filter) (bool, error) {
	local, err := ld.local.FindOne(ctx, filter)
	if err != nil {
		return false, err
	}
	if local == nil {
		return false, nil
	}

	if _, err := ld.table.Free(local.GetID()); err != nil {
		return false, err
	}
	return ld.local.DeleteOne(ctx, storage.Where[ulid.ULID](scheme.KeyID).EQ(local.GetID()))
}

func (ld *Loader) unloadMany(ctx context.Context, filter *storage.Filter) (int, error) {
	locals, err := ld.local.FindMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	for _, local := range locals {
		if _, err := ld.table.Free(local.GetID()); err != nil {
			return 0, err
		}
	}

	var ids []ulid.ULID
	for _, local := range locals {
		ids = append(ids, local.GetID())
	}
	return ld.local.DeleteMany(ctx, storage.Where[ulid.ULID](scheme.KeyID).IN(ids...))
}
