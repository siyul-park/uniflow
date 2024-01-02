package memdb

import (
	"context"
	"sync"

	"github.com/emirpasic/gods/containers"
	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/sets"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type (
	IndexView struct {
		names  []string
		models []database.IndexModel
		data   []maps.Map
		lock   sync.RWMutex
	}
)

var _ database.IndexView = &IndexView{}

var (
	keyID = primitive.NewString("id")
)

var (
	ErrIndexConflict   = errors.New("index is conflict")
	ErrIndexNotFound   = errors.New("index is not found")
	ErrInvalidDocument = errors.New("document is invalid")
)

func NewIndexView() *IndexView {
	v := &IndexView{
		names:  nil,
		models: nil,
		data:   nil,
		lock:   sync.RWMutex{},
	}

	primaryModel := database.IndexModel{
		Keys:    []string{"id"},
		Name:    "_id",
		Unique:  true,
		Partial: nil,
	}

	v.names = append(v.names, primaryModel.Name)
	v.models = append(v.models, primaryModel)
	v.data = append(v.data, treemap.NewWith(comparator))

	return v
}

func (v *IndexView) List(_ context.Context) ([]database.IndexModel, error) {
	v.lock.RLock()
	defer v.lock.RUnlock()

	return v.models, nil
}

func (v *IndexView) Create(_ context.Context, index database.IndexModel) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	name := index.Name

	for i, n := range v.names {
		if n == name {
			v.names = append(v.names[:i], v.names[i+1:]...)
			v.models = append(v.models[:i], v.models[i+1:]...)
			v.data = append(v.data[:i], v.data[i+1:]...)
		}
	}

	v.names = append(v.names, name)
	v.models = append(v.models, index)
	v.data = append(v.data, treemap.NewWith(comparator))

	return nil
}

func (v *IndexView) Drop(_ context.Context, name string) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	for i, n := range v.names {
		if n == name {
			v.names = append(v.names[:i], v.names[i+1:]...)
			v.models = append(v.models[:i], v.models[i+1:]...)
			v.data = append(v.data[:i], v.data[i+1:]...)
		}
	}

	return nil
}

func (v *IndexView) insertMany(ctx context.Context, docs []*primitive.Map) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	for i, doc := range docs {
		if err := v.insertOne(ctx, doc); err != nil {
			for i--; i >= 0; i-- {
				_ = v.deleteOne(ctx, doc)
			}
			return err
		}
	}
	return nil
}

func (v *IndexView) deleteMany(ctx context.Context, docs []*primitive.Map) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	for i, doc := range docs {
		if err := v.deleteOne(ctx, doc); err != nil {
			for ; i >= 0; i-- {
				_ = v.insertOne(ctx, doc)
			}
			return err
		}
	}
	return nil
}

func (v *IndexView) deleteAll(_ context.Context) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.data = nil

	return nil
}

func (v *IndexView) insertOne(ctx context.Context, doc *primitive.Map) error {
	id, ok := doc.Get(keyID)
	if !ok {
		return errors.WithStack(ErrIndexConflict)
	}

	for i, model := range v.models {
		match := ParseFilter(model.Partial)
		if !match(doc) {
			continue
		}

		if err := func() error {
			cur := v.data[i]

			for i, k := range model.Keys {
				value, _ := primitive.Pick[primitive.Value](doc, k)
				child, ok := cur.Get(value)

				if i < len(model.Keys)-1 {
					if !ok {
						child = treemap.NewWith(comparator)
						cur.Put(value, child)
					}
					cur = child.(maps.Map)
				} else if model.Unique {
					if !ok {
						cur.Put(value, id)
					} else if child != id {
						return ErrIndexConflict
					}
				} else {
					if !ok {
						child = treeset.NewWith(comparator)
						cur.Put(value, child)
					}
					child.(sets.Set).Add(id)
				}
			}

			return nil
		}(); err != nil {
			_ = v.deleteOne(ctx, doc)
			return err
		}
	}

	return nil
}

func (v *IndexView) deleteOne(_ context.Context, doc *primitive.Map) error {
	id, ok := doc.Get(keyID)
	if !ok {
		return nil
	}

	for i, model := range v.models {
		match := ParseFilter(model.Partial)
		if !match(doc) {
			continue
		}

		if err := func() error {
			cur := v.data[i]

			var nodes []containers.Container
			nodes = append(nodes, cur)

			var keys []primitive.Value
			keys = append(keys, nil)

			for i, k := range model.Keys {
				value, _ := primitive.Pick[primitive.Value](doc, k)
				child, ok := cur.Get(value)
				if !ok {
					return nil
				}

				if i < len(model.Keys)-1 {
					cur = child.(maps.Map)

					nodes = append(nodes, cur)
					keys = append(keys, value)
				} else if model.Unique {
					if primitive.Compare(id, child.(primitive.Value)) == 0 {
						cur.Remove(value)
					}
				} else {
					nodes = append(nodes, child.(sets.Set))
					keys = append(keys, value)
					child.(sets.Set).Remove(id)
				}
			}

			for i := len(nodes) - 1; i >= 0; i-- {
				child := nodes[i]

				if child.Empty() && i > 0 {
					parent := nodes[i-1]
					key := keys[i]

					if p, ok := parent.(maps.Map); ok {
						p.Remove(key)
					}
				}
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}
