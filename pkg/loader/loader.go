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
		n, err := ld.table.Insert(n)
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

		if err := ld.resolveLinks(ctx, local, remote); err != nil {
			return nil, err
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
			if sym, err := ld.table.Insert(n); err != nil {
				return nil, err
			} else {
				nodes = append(nodes, sym)
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

	for id, remote := range idToRemote {
		local := idToLocal[id]
		if err := ld.resolveLinks(ctx, local, remote); err != nil {
			return nil, err
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

	if err := ld.resolveLinks(ctx, local, nil); err != nil {
		return false, err
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
		if err := ld.resolveLinks(ctx, local, nil); err != nil {
			return 0, err
		}
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

func (ld *Loader) resolveLinks(ctx context.Context, local scheme.Spec, remote scheme.Spec) error {
	var n node.Node
	var ok bool

	var spec scheme.Spec
	var localLinks links
	var remoteLinks links

	if local != nil {
		spec = local
		localLinks = local.GetLinks()
		n, ok = ld.table.Lookup(local.GetID())
	}
	if remote != nil {
		spec = remote
		remoteLinks = remote.GetLinks()
		if !ok {
			n, ok = ld.table.Lookup(remote.GetID())
		}
	}
	if !ok {
		return nil
	}

	deletions := localLinks
	additions := remoteLinks

	undefined := links{}

	for name, locations := range deletions {
		for _, location := range locations {
			id := location.ID

			if id == (ulid.ULID{}) {
				if location.Name != "" {
					filter := storage.Where[string](scheme.KeyNamespace).EQ(spec.GetNamespace())
					filter = filter.And(storage.Where[string](scheme.KeyName).EQ(location.Name))
					if spec, err := ld.local.FindOne(ctx, filter); err != nil {
						return err
					} else if spec != nil {
						id = spec.GetID()
					}
				}
			}

			if id != (ulid.ULID{}) {
				if ref, ok := ld.table.Lookup(id); ok {
					referenced := ld.referenced[ref.ID()]
					var locations []scheme.PortLocation
					for _, location := range referenced[location.Port] {
						if location.ID != n.ID() || location.Port != name {
							locations = append(locations, location)
						}
					}
					if len(locations) > 0 {
						referenced[location.Port] = locations
						ld.referenced[ref.ID()] = referenced
					} else if referenced != nil {
						delete(referenced, location.Port)
						ld.referenced[ref.ID()] = referenced
					}
				}
			}
		}
	}

	for name, locations := range additions {
		p1, ok := n.Port(name)
		if !ok {
			undefined[name] = locations
			continue
		}

		for _, location := range locations {
			filter := storage.Where[string](scheme.KeyNamespace).EQ(spec.GetNamespace())
			if location.ID != (ulid.ULID{}) {
				filter = filter.And(storage.Where[ulid.ULID](scheme.KeyID).EQ(location.ID))
			} else if location.Name != "" {
				filter = filter.And(storage.Where[string](scheme.KeyName).EQ(location.Name))
			} else {
				continue
			}

			// TODO: use load many
			if ref, err := ld.loadOne(ctx, filter); err != nil {
				return err
			} else if ref != nil {
				if p2, ok := ref.Port(location.Port); ok {
					p1.Link(p2)

					referenced := ld.referenced[ref.ID()]
					if referenced == nil {
						referenced = links{}
					}
					referenced[location.Port] = append(referenced[location.Port], scheme.PortLocation{
						ID:   n.ID(),
						Port: name,
					})
					ld.referenced[ref.ID()] = referenced
				} else {
					undefined[name] = append(undefined[name], location)
				}
			} else {
				undefined[name] = append(undefined[name], location)
			}
		}
	}

	undefined = diffLinks(unionLinks(ld.undefined[n.ID()], undefined), deletions)

	if len(undefined) > 0 {
		ld.undefined[n.ID()] = undefined
	} else {
		delete(ld.undefined, n.ID())
	}

	if remote == nil {
		ld.removeReference(ctx, n.ID())
	} else {
		for name, locations := range ld.referenced[spec.GetID()] {
			p1, ok := n.Port(name)
			if !ok {
				continue
			}
			for _, location := range locations {
				if ref, ok := ld.table.Lookup(location.ID); ok {
					if p2, ok := ref.Port(location.Port); ok {
						p1.Link(p2)
					}
				}
			}
		}

		for id, additions := range ld.undefined {
			if ref, err := ld.local.FindOne(ctx, storage.Where[ulid.ULID](scheme.KeyID).EQ(id)); err != nil {
				return err
			} else if ref == nil {
				ld.removeReference(ctx, id)
				delete(ld.undefined, id)
				continue
			} else if ref.GetNamespace() != spec.GetNamespace() {
				continue
			}

			undefined := make(links, len(additions))

			if ref, ok := ld.table.Lookup(id); ok {
				for name, locations := range additions {
					p1, ok := ref.Port(name)
					if !ok {
						continue
					}

					for _, location := range locations {
						if (location.ID == spec.GetID()) || (location.Name != "" && location.Name == spec.GetName()) {
							if p2, ok := n.Port(location.Port); ok {
								p1.Link(p2)

								referenced := ld.referenced[n.ID()]
								if referenced == nil {
									referenced = links{}
								}
								referenced[location.Port] = append(referenced[location.Port], scheme.PortLocation{
									ID:   ref.ID(),
									Port: name,
								})
								ld.referenced[n.ID()] = referenced
							} else {
								undefined[name] = append(undefined[name], location)
							}
						} else {
							undefined[name] = append(undefined[name], location)
						}
					}
				}
			}

			ld.undefined[id] = undefined
		}
	}

	return nil
}

func (ld *Loader) removeReference(ctx context.Context, id ulid.ULID) {
	for name, locations := range ld.referenced[id] {
		for _, location := range locations {
			if ref, ok := ld.table.Lookup(location.ID); ok {
				undefined := ld.undefined[ref.ID()]
				if undefined == nil {
					undefined = links{}
				}
				undefined[location.Port] = append(undefined[location.Port], scheme.PortLocation{
					ID:   id,
					Port: name,
				})
				ld.undefined[ref.ID()] = undefined
			}
		}
	}
	delete(ld.referenced, id)
}

func diffLinks(l1 links, l2 links) links {
	diff := make(links, len(l1))
	for name, locations1 := range l1 {
		diffLocationSet := map[scheme.PortLocation]struct{}{}
		for _, location := range locations1 {
			diffLocationSet[location] = struct{}{}
		}
		if locations2, ok := l2[name]; ok {
			for _, location := range locations2 {
				delete(diffLocationSet, location)
			}
		}

		var diffLocations []scheme.PortLocation
		for location := range diffLocationSet {
			diffLocations = append(diffLocations, location)
		}

		if len(diffLocations) > 0 {
			diff[name] = diffLocations
		}
	}

	if len(diff) == 0 {
		return nil
	}
	return diff
}

func unionLinks(l1 links, l2 links) links {
	unionSet := make(map[string]map[scheme.PortLocation]struct{}, len(l1)+len(l2))
	for name, locations := range l1 {
		unionLocationSet := map[scheme.PortLocation]struct{}{}
		for _, location := range locations {
			unionLocationSet[location] = struct{}{}
		}
		unionSet[name] = unionLocationSet
	}
	for name, locations := range l2 {
		unionLocationSet := unionSet[name]
		if len(unionLocationSet) == 0 {
			unionLocationSet = map[scheme.PortLocation]struct{}{}
		}
		for _, location := range locations {
			unionLocationSet[location] = struct{}{}
		}
		unionSet[name] = unionLocationSet
	}

	union := make(links, len(unionSet))
	for name, locationSet := range unionSet {
		var locations []scheme.PortLocation
		for location := range locationSet {
			locations = append(locations, location)
		}

		union[name] = locations
	}

	return union
}
