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
	iv := &IndexView{
		names:  nil,
		models: nil,
		data:   nil,
		lock:   sync.RWMutex{},
	}
	_ = iv.Create(context.Background(), database.IndexModel{
		Keys:    []string{"id"},
		Name:    "_id",
		Unique:  true,
		Partial: nil,
	})

	return iv
}

func (iv *IndexView) List(_ context.Context) ([]database.IndexModel, error) {
	iv.lock.RLock()
	defer iv.lock.RUnlock()

	return iv.models, nil
}

func (iv *IndexView) Create(_ context.Context, index database.IndexModel) error {
	iv.lock.Lock()
	defer iv.lock.Unlock()

	name := index.Name

	for i, n := range iv.names {
		if n == name {
			iv.names = append(iv.names[:i], iv.names[i+1:]...)
			iv.models = append(iv.models[:i], iv.models[i+1:]...)
			iv.data = append(iv.data[:i], iv.data[i+1:]...)
		}
	}

	iv.names = append(iv.names, name)
	iv.models = append(iv.models, index)
	iv.data = append(iv.data, treemap.NewWith(comparator))

	return nil
}

func (iv *IndexView) Drop(_ context.Context, name string) error {
	iv.lock.Lock()
	defer iv.lock.Unlock()

	for i, n := range iv.names {
		if n == name {
			iv.names = append(iv.names[:i], iv.names[i+1:]...)
			iv.models = append(iv.models[:i], iv.models[i+1:]...)
			iv.data = append(iv.data[:i], iv.data[i+1:]...)
		}
	}

	return nil
}

func (iv *IndexView) insertMany(ctx context.Context, docs []*primitive.Map) error {
	iv.lock.Lock()
	defer iv.lock.Unlock()

	for i, doc := range docs {
		if err := iv.insertOne(ctx, doc); err != nil {
			for i--; i >= 0; i-- {
				_ = iv.deleteOne(ctx, doc)
			}
			return err
		}
	}
	return nil
}

func (iv *IndexView) deleteMany(ctx context.Context, docs []*primitive.Map) error {
	iv.lock.Lock()
	defer iv.lock.Unlock()

	for i, doc := range docs {
		if err := iv.deleteOne(ctx, doc); err != nil {
			for ; i >= 0; i-- {
				_ = iv.insertOne(ctx, doc)
			}
			return err
		}
	}
	return nil
}

func (iv *IndexView) deleteAll(_ context.Context) error {
	iv.lock.Lock()
	defer iv.lock.Unlock()

	iv.data = nil

	return nil
}

func (iv *IndexView) findMany(_ context.Context, examples []*primitive.Map) ([]primitive.Object, error) {
	iv.lock.RLock()
	defer iv.lock.RUnlock()

	ids := treeset.NewWith(comparator)

	for _, example := range examples {
		if err := func() error {
			for i, model := range iv.models {
				curr := iv.data[i]

				visits := make(map[string]bool, example.Len())
				for _, k := range example.Keys() {
					if k, ok := k.(primitive.String); ok {
						visits[k.String()] = false
					} else {
						return ErrInvalidDocument
					}
				}
				next := false

				var i int
				var k string
				for i, k = range model.Keys {
					if obj, ok := primitive.Pick[primitive.Object](example, k); ok {
						visits[k] = true
						if sub, ok := curr.Get(obj); ok {
							if i < len(model.Keys)-1 {
								curr = sub.(maps.Map)
							} else {
								if model.Unique {
									ids.Add(sub)
								} else {
									ids.Add(sub.(sets.Set).Values()...)
									return nil
								}
							}
						} else {
							next = true
							break
						}
					} else {
						break
					}
				}

				for _, v := range visits {
					if !v {
						next = true
					}
				}
				if next {
					continue
				}

				var parent []maps.Map
				parent = append(parent, curr)

				depth := len(model.Keys) - 1
				if !model.Unique {
					depth += 1
				}

				for ; i < depth; i++ {
					var children []maps.Map
					for _, curr := range parent {
						for _, v := range curr.Values() {
							children = append(children, v.(maps.Map))
						}
					}
					parent = children
				}

				for _, curr := range parent {
					ids.Add(curr.Values()...)
				}

				return nil
			}

			return ErrIndexNotFound
		}(); err != nil {
			return nil, err
		}
	}

	var uniqueIds []primitive.Object
	for _, v := range ids.Values() {
		uniqueIds = append(uniqueIds, v.(primitive.Object))
	}
	return uniqueIds, nil
}

func (iv *IndexView) insertOne(ctx context.Context, doc *primitive.Map) error {
	id, ok := doc.Get(keyID)
	if !ok {
		return ErrIndexConflict
	}

	for i, model := range iv.models {
		if err := func() error {
			curr := iv.data[i]

			if !ParseFilter(model.Partial)(doc) {
				return nil
			}

			for i, k := range model.Keys {
				obj, _ := primitive.Pick[primitive.Object](doc, k)

				if i < len(model.Keys)-1 {
					sub, ok := curr.Get(obj)
					if !ok {
						sub = treemap.NewWith(comparator)
						curr.Put(obj, sub)
					}
					curr = sub.(maps.Map)
				} else if model.Unique {
					if r, ok := curr.Get(obj); !ok {
						curr.Put(obj, id)
					} else if r != id {
						return ErrIndexConflict
					}
				} else {
					r, ok := curr.Get(obj)
					if !ok {
						r = treeset.NewWith(comparator)
						curr.Put(obj, r)
					}
					r.(sets.Set).Add(id)
				}
			}

			return nil
		}(); err != nil {
			_ = iv.deleteOne(ctx, doc)
			return err
		}
	}

	return nil
}

func (iv *IndexView) deleteOne(_ context.Context, doc *primitive.Map) error {
	id, ok := doc.Get(keyID)
	if !ok {
		return nil
	}

	for i, model := range iv.models {
		if err := func() error {
			curr := iv.data[i]

			if !ParseFilter(model.Partial)(doc) {
				return nil
			}

			var nodes []containers.Container
			nodes = append(nodes, curr)
			var keys []primitive.Object
			keys = append(keys, nil)

			for i, k := range model.Keys {
				obj, _ := primitive.Pick[primitive.Object](doc, k)

				if i < len(model.Keys)-1 {
					if sub, ok := curr.Get(obj); ok {
						curr = sub.(maps.Map)

						nodes = append(nodes, curr)
						keys = append(keys, obj)
					} else {
						return nil
					}
				} else if model.Unique {
					if r, ok := curr.Get(obj); ok && primitive.Equal(id, r.(primitive.Object)) {
						curr.Remove(obj)
					}
				} else {
					if r, ok := curr.Get(obj); ok {
						nodes = append(nodes, r.(sets.Set))
						keys = append(keys, obj)
						r.(sets.Set).Remove(id)
					}
				}
			}

			for i := len(nodes) - 1; i >= 0; i-- {
				node := nodes[i]

				if node.Empty() && i > 0 {
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
