package memdb

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/internal/pool"
	"github.com/siyul-park/uniflow/internal/util"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type (
	IndexView struct {
		names  []string
		models []database.IndexModel
		data   []*sync.Map
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
	iv.data = append(iv.data, pool.GetMap())

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

	ids := pool.GetMap()
	defer pool.PutMap(ids)

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
					if obj, ok := primitive.Get[any](example, k); ok {
						v := primitive.Interface(obj)

						hash, err := util.Hash(v)
						if err != nil {
							return err
						}
						visits[k] = true
						if sub, ok := curr.Load(hash); ok {
							if i < len(model.Keys)-1 {
								curr = sub.(*sync.Map)
							} else {
								if model.Unique {
									if hsub, err := util.Hash(sub); err != nil {
										return err
									} else {
										ids.Store(hsub, sub)
										return nil
									}
								} else {
									sub.(*sync.Map).Range(func(key, val any) bool {
										ids.Store(key, val)
										return true
									})
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

				var parent []*sync.Map
				parent = append(parent, curr)

				depth := len(model.Keys) - 1
				if !model.Unique {
					depth += 1
				}

				for ; i < depth; i++ {
					var children []*sync.Map
					for _, curr := range parent {
						curr.Range(func(_, value any) bool {
							children = append(children, value.(*sync.Map))
							return true
						})
					}
					parent = children
				}

				for _, curr := range parent {
					curr.Range(func(k, v any) bool {
						ids.Store(k, v)
						return true
					})
				}

				return nil
			}

			return ErrIndexNotFound
		}(); err != nil {
			return nil, err
		}
	}

	var uniqueIds []primitive.Object
	ids.Range(func(_, val any) bool {
		uniqueIds = append(uniqueIds, val.(primitive.Object))
		return true
	})
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
				obj, _ := primitive.Get[primitive.Object](doc, k)
				v := primitive.Interface(obj)

				hash, err := util.Hash(v)
				if err != nil {
					return err
				}
				if i < len(model.Keys)-1 {
					cm := pool.GetMap()
					sub, load := curr.LoadOrStore(hash, cm)
					if load {
						pool.PutMap(cm)
					}
					curr = sub.(*sync.Map)
				} else if model.Unique {
					if r, loaded := curr.LoadOrStore(hash, id); loaded && r != id {
						return ErrIndexConflict
					}
				} else {
					cm := pool.GetMap()
					r, load := curr.LoadOrStore(hash, cm)
					if load {
						pool.PutMap(cm)
					}
					r.(*sync.Map).Store(hash, id)
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

	hid, err := util.Hash(id)
	if err != nil {
		return err
	}

	for i, model := range iv.models {
		if err := func() error {
			curr := iv.data[i]

			if !ParseFilter(model.Partial)(doc) {
				return nil
			}

			var nodes []*sync.Map
			nodes = append(nodes, curr)
			var keys []any
			keys = append(keys, nil)

			for i, k := range model.Keys {
				obj, _ := primitive.Get[primitive.Object](doc, k)
				v := primitive.Interface(obj)

				hash, err := util.Hash(v)
				if err != nil {
					return err
				}

				if i < len(model.Keys)-1 {
					if sub, ok := curr.Load(hash); ok {
						curr = sub.(*sync.Map)

						nodes = append(nodes, curr)
						keys = append(keys, hash)
					} else {
						return nil
					}
				} else if model.Unique {
					if r, loaded := curr.Load(hash); loaded && primitive.Equal(id, r.(primitive.Object)) {
						curr.Delete(hash)
					}
				} else {
					if r, loaded := curr.Load(hash); loaded {
						nodes = append(nodes, r.(*sync.Map))
						keys = append(keys, hash)
						r.(*sync.Map).Delete(hid)
					}
				}
			}

			for i := len(nodes) - 1; i >= 0; i-- {
				node := nodes[i]

				empty := true
				node.Range(func(_, _ any) bool {
					empty = false
					return false
				})

				if empty && i > 0 {
					parent := nodes[i-1]
					key := keys[i]

					parent.Delete(key)
					pool.PutMap(node)
				}
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}
