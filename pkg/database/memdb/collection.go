package memdb

import (
	"context"
	"sort"
	"sync"

	"github.com/emirpasic/gods/maps"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type (
	Collection struct {
		name      string
		data      maps.Map
		indexView *IndexView
		streams   []*Stream
		matches   []func(*primitive.Map) bool
		mu        sync.RWMutex
	}

	internalEvent struct {
		database.Event
		Document *primitive.Map
	}
)

var _ database.Collection = &Collection{}

var (
	ErrPKNotFound   = errors.New("primary key is not found")
	ErrPKDuplicated = errors.New("primary key is duplicated")
)

func NewCollection(name string) *Collection {
	return &Collection{
		name:      name,
		data:      treemap.NewWith(comparator),
		indexView: NewIndexView(),
		mu:        sync.RWMutex{},
	}
}

func (coll *Collection) Name() string {
	coll.mu.RLock()
	defer coll.mu.RUnlock()

	return coll.name
}

func (coll *Collection) Indexes() database.IndexView {
	coll.mu.RLock()
	defer coll.mu.RUnlock()

	return coll.indexView
}

func (coll *Collection) Watch(ctx context.Context, filter *database.Filter) (database.Stream, error) {
	coll.mu.Lock()
	defer coll.mu.Unlock()

	stream := NewStream()

	coll.streams = append(coll.streams, stream)
	coll.matches = append(coll.matches, ParseFilter(filter))

	go func() {
		select {
		case <-stream.Done():
			coll.unwatch(stream)
		case <-ctx.Done():
			_ = stream.Close()
			coll.unwatch(stream)
		}
	}()

	return stream, nil
}

func (coll *Collection) InsertOne(ctx context.Context, doc *primitive.Map) (primitive.Value, error) {
	if id, err := coll.insertOne(ctx, doc); err != nil {
		return nil, err
	} else {
		coll.emit(internalEvent{
			Event: database.Event{
				OP:         database.EventInsert,
				DocumentID: id,
			},
			Document: doc,
		})
		return id, nil
	}
}

func (coll *Collection) InsertMany(ctx context.Context, docs []*primitive.Map) ([]primitive.Value, error) {
	if ids, err := coll.insertMany(ctx, docs); err != nil {
		return nil, err
	} else {
		for i, doc := range docs {
			coll.emit(internalEvent{
				Event: database.Event{
					OP:         database.EventInsert,
					DocumentID: ids[i],
				},
				Document: doc,
			})
		}
		return ids, nil
	}
}

func (coll *Collection) UpdateOne(ctx context.Context, filter *database.Filter, patch *primitive.Map, opts ...*database.UpdateOptions) (bool, error) {
	opt := database.MergeUpdateOptions(opts)

	upsert := false
	if opt != nil && opt.Upsert != nil {
		upsert = lo.FromPtr(opt.Upsert)
	}

	old, err := coll.findOne(ctx, filter)
	if err != nil {
		return false, err
	}
	if old == nil && !upsert {
		return false, nil
	}

	var id primitive.Value
	if old != nil {
		id = old.GetOr(keyID, nil)
	}
	if id == nil {
		id = patch.GetOr(keyID, nil)
	}
	if id == nil {
		if examples, ok := FilterToExample(filter); ok {
			for _, example := range examples {
				if v, ok := example.Get(keyID); ok {
					if id == nil {
						id = v
					} else {
						return false, errors.Wrap(errors.WithStack(ErrPKDuplicated), database.ErrCodeWrite)
					}
				}
			}
		}
	}
	if id == nil {
		return false, errors.Wrap(errors.WithStack(ErrPKNotFound), database.ErrCodeWrite)
	}

	if old != nil {
		if _, err := coll.deleteOne(ctx, old); err != nil {
			return false, err
		}
	}

	doc := patch
	if _, ok := doc.Get(keyID); !ok {
		doc = doc.Set(keyID, id)
	}

	if _, err := coll.insertOne(ctx, doc); err != nil {
		_, _ = coll.InsertOne(ctx, old)
		return false, err
	}

	coll.emit(internalEvent{
		Event: database.Event{
			OP:         database.EventUpdate,
			DocumentID: id,
		},
		Document: doc,
	})

	return true, nil
}

func (coll *Collection) UpdateMany(ctx context.Context, filter *database.Filter, patch *primitive.Map, opts ...*database.UpdateOptions) (int, error) {
	opt := database.MergeUpdateOptions(opts)
	upsert := false
	if opt != nil && opt.Upsert != nil {
		upsert = lo.FromPtr(opt.Upsert)
	}

	old, err := coll.findMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	if len(old) == 0 {
		if !upsert {
			return 0, nil
		}

		id := patch.GetOr(keyID, nil)
		if id == nil {
			if examples, ok := FilterToExample(filter); ok {
				for _, example := range examples {
					if v, ok := example.Get(keyID); ok {
						if id == nil {
							id = v
						} else {
							return 0, errors.Wrap(errors.WithStack(ErrPKDuplicated), database.ErrCodeWrite)
						}
					}
				}
			}
		}

		doc := patch
		if _, ok := doc.Get(keyID); !ok {
			doc = doc.Set(keyID, id)
		}
		if _, err := coll.insertOne(ctx, doc); err != nil {
			return 0, err
		}
		return 1, nil
	}

	if _, err := coll.deleteMany(ctx, old); err != nil {
		return 0, err
	}

	docs := make([]*primitive.Map, len(old))
	for i, doc := range old {
		doc = patch.Set(keyID, doc.GetOr(keyID, nil))
		docs[i] = doc
	}

	ids, err := coll.insertMany(ctx, docs)
	if err != nil {
		_, _ = coll.insertMany(ctx, old)
		return 0, err
	}

	for i, doc := range docs {
		coll.emit(internalEvent{
			Event: database.Event{
				OP:         database.EventUpdate,
				DocumentID: ids[i],
			},
			Document: doc,
		})
	}

	return len(docs), nil
}

func (coll *Collection) DeleteOne(ctx context.Context, filter *database.Filter) (bool, error) {
	if doc, err := coll.findOne(ctx, filter); err != nil {
		return false, err
	} else if doc, err := coll.deleteOne(ctx, doc); err != nil {
		return false, err
	} else {
		if doc != nil {
			if id, ok := doc.Get(keyID); ok {
				coll.emit(internalEvent{
					Event: database.Event{
						OP:         database.EventDelete,
						DocumentID: id,
					},
					Document: doc,
				})
			}
		}
		return doc != nil, nil
	}
}

func (coll *Collection) DeleteMany(ctx context.Context, filter *database.Filter) (int, error) {
	if docs, err := coll.findMany(ctx, filter); err != nil {
		return 0, err
	} else if docs, err := coll.deleteMany(ctx, docs); err != nil {
		return 0, err
	} else {
		for _, doc := range docs {
			if id, ok := doc.Get(keyID); ok {
				coll.emit(internalEvent{
					Event: database.Event{
						OP:         database.EventDelete,
						DocumentID: id,
					},
					Document: doc,
				})
			}
		}
		return len(docs), nil
	}
}

func (coll *Collection) FindOne(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) (*primitive.Map, error) {
	return coll.findOne(ctx, filter, opts...)
}

func (coll *Collection) FindMany(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) ([]*primitive.Map, error) {
	return coll.findMany(ctx, filter, opts...)
}

func (coll *Collection) Drop(ctx context.Context) error {
	data, err := func() (maps.Map, error) {
		coll.mu.Lock()
		defer coll.mu.Unlock()

		data := coll.data
		coll.data = treemap.NewWith(comparator)

		if err := coll.indexView.deleteAll(ctx); err != nil {
			return nil, err
		}

		return data, nil
	}()
	if err != nil {
		return err
	}

	for _, val := range data.Values() {
		doc := val.(*primitive.Map)
		if id, ok := doc.Get(keyID); ok {
			coll.emit(internalEvent{
				Event: database.Event{
					OP:         database.EventDelete,
					DocumentID: id,
				},
				Document: doc,
			})
		}
	}

	coll.mu.Lock()
	defer coll.mu.Unlock()

	for _, s := range coll.streams {
		if err := s.Close(); err != nil {
			return err
		}
	}
	coll.streams = nil

	return nil
}

func (coll *Collection) insertOne(ctx context.Context, doc *primitive.Map) (primitive.Value, error) {
	if ids, err := coll.insertMany(ctx, []*primitive.Map{doc}); err != nil {
		return nil, err
	} else {
		return ids[0], nil
	}
}

func (coll *Collection) insertMany(ctx context.Context, docs []*primitive.Map) ([]primitive.Value, error) {
	coll.mu.Lock()
	defer coll.mu.Unlock()

	ids := make([]primitive.Value, len(docs))
	for i, doc := range docs {
		if id, ok := doc.Get(keyID); !ok {
			return nil, errors.Wrap(errors.WithStack(ErrPKNotFound), database.ErrCodeWrite)
		} else if _, ok := coll.data.Get(id); ok {
			return nil, errors.Wrap(errors.WithStack(ErrPKDuplicated), database.ErrCodeWrite)
		} else {
			ids[i] = id
		}
	}

	if err := coll.indexView.insertMany(ctx, docs); err != nil {
		return nil, errors.Wrap(err, database.ErrCodeWrite)
	}
	for i, doc := range docs {
		coll.data.Put(ids[i], doc)
	}

	return ids, nil
}

func (coll *Collection) findOne(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) (*primitive.Map, error) {
	opt := database.MergeFindOptions(append(opts, lo.ToPtr(database.FindOptions{Limit: lo.ToPtr(1)})))

	if docs, err := coll.findMany(ctx, filter, opt); err != nil {
		return nil, err
	} else if len(docs) > 0 {
		return docs[0], nil
	} else {
		return nil, nil
	}
}

func (coll *Collection) findMany(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) ([]*primitive.Map, error) {
	coll.mu.RLock()
	defer coll.mu.RUnlock()

	opt := database.MergeFindOptions(opts)

	limit := -1
	skip := 0
	var sorts []database.Sort

	if opt != nil {
		if opt.Limit != nil {
			limit = lo.FromPtr(opt.Limit)
		}
		if opt.Skip != nil {
			skip = lo.FromPtr(opt.Skip)
		}
		if opt.Sorts != nil {
			sorts = opt.Sorts
		}
	}

	match := ParseFilter(filter)

	var docs []*primitive.Map
	for _, key := range coll.data.Keys() {
		value, _ := coll.data.Get(key)
		if len(sorts) == 0 && limit == len(docs) {
			continue
		}
		if match(value.(*primitive.Map)) {
			docs = append(docs, value.(*primitive.Map))
		}
	}

	if skip >= len(docs) {
		return nil, nil
	}

	if len(sorts) > 0 {
		compare := ParseSorts(sorts)
		sort.Slice(docs, func(i, j int) bool {
			return compare(docs[i], docs[j])
		})
	}

	if limit >= 0 {
		if len(docs) > limit+skip {
			docs = docs[skip : limit+skip]
		} else {
			docs = docs[skip:]
		}
	}

	return docs, nil
}

func (coll *Collection) deleteOne(ctx context.Context, doc *primitive.Map) (*primitive.Map, error) {
	if docs, err := coll.deleteMany(ctx, []*primitive.Map{doc}); err != nil {
		return nil, err
	} else if len(docs) > 0 {
		return docs[0], nil
	} else {
		return nil, nil
	}
}

func (coll *Collection) deleteMany(ctx context.Context, docs []*primitive.Map) ([]*primitive.Map, error) {
	coll.mu.Lock()
	defer coll.mu.Unlock()

	ids := make([]primitive.Value, 0, len(docs))
	deletes := make([]*primitive.Map, 0, len(docs))
	for _, doc := range docs {
		if doc == nil {
			continue
		}
		if id, ok := doc.Get(keyID); !ok {
			continue
		} else {
			ids = append(ids, id)
			deletes = append(deletes, doc)
		}
	}

	if err := coll.indexView.deleteMany(ctx, deletes); err != nil {
		return nil, errors.Wrap(err, database.ErrCodeDelete)
	}

	for _, id := range ids {
		coll.data.Remove(id)
	}

	return deletes, nil
}

func (coll *Collection) unwatch(stream database.Stream) {
	coll.mu.Lock()
	defer coll.mu.Unlock()

	for i, s := range coll.streams {
		if s == stream {
			coll.streams = append(coll.streams[:i], coll.streams[i+1:]...)
			coll.matches = append(coll.matches[:i], coll.matches[i+1:]...)
			return
		}
	}
}

func (coll *Collection) emit(event internalEvent) {
	coll.mu.RLock()
	defer coll.mu.RUnlock()

	for i, s := range coll.streams {
		if coll.matches[i](event.Document) {
			s.Emit(event.Event)
		}
	}
}
