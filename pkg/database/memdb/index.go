package memdb

import (
	"context"
	"sync"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type IndexView struct {
	names  []string
	models []database.IndexModel
	data   []maps.Map
	lock   sync.RWMutex
}

var _ database.IndexView = &IndexView{}

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

func (v *IndexView) insertMany(docs []*primitive.Map) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	for i, doc := range docs {
		if err := v.insertOne(doc); err != nil {
			for i--; i >= 0; i-- {
				v.deleteOne(doc)
			}
			return err
		}
	}
	return nil
}

func (v *IndexView) deleteMany(docs []*primitive.Map) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	for _, doc := range docs {
		v.deleteOne(doc)
	}
	return nil
}

func (v *IndexView) insertOne(doc *primitive.Map) error {
	id, ok := doc.Get(keyID)
	if !ok {
		return errors.WithStack(ErrIndexConflict)
	}

	for i, model := range v.models {
		match := parseFilter(model.Partial)
		if !match(doc) {
			continue
		}

		cur := v.data[i]

		for i, k := range model.Keys {
			value, _ := primitive.Pick[primitive.Value](doc, k)

			c, _ := cur.Get(value)
			child, ok := c.(maps.Map)
			if !ok {
				child = treemap.NewWith(comparator)
				cur.Put(value, child)
			}

			if i < len(model.Keys)-1 {
				cur = child
			} else {
				child.Put(id, nil)

				if model.Unique && child.Size() > 1 {
					child.Remove(id)
					v.deleteOne(doc)
					return errors.WithStack(ErrIndexConflict)
				}
			}
		}
	}

	return nil
}

func (v *IndexView) deleteOne(doc *primitive.Map) {
	id, ok := doc.Get(keyID)
	if !ok {
		return
	}

	for i, model := range v.models {
		match := parseFilter(model.Partial)
		if !match(doc) {
			continue
		}

		cur := v.data[i]

		var nodes []maps.Map
		nodes = append(nodes, cur)

		var keys []primitive.Value
		keys = append(keys, nil)

		for i, k := range model.Keys {
			value, _ := primitive.Pick[primitive.Value](doc, k)

			c, _ := cur.Get(value)
			child, ok := c.(maps.Map)
			if !ok {
				nodes = nil
				keys = nil

				break
			}

			nodes = append(nodes, child)
			keys = append(keys, value)

			if i < len(model.Keys)-1 {
				cur = child
			} else {
				child.Remove(id)
			}
		}

		for i := len(nodes) - 1; i >= 0; i-- {
			child := nodes[i]

			if child.Empty() && i > 0 {
				parent := nodes[i-1]
				key := keys[i]

				parent.Remove(key)
			}
		}
	}
}

func (v *IndexView) dropData() {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.data = nil
}
