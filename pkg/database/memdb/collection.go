package memdb

import (
	"context"
	"sort"
	"sync"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/internal/pool"
	"github.com/siyul-park/uniflow/internal/util"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type (
	Collection struct {
		name          string
		data          maps.Map
		indexView     *IndexView
		streams       []*Stream
		streamMatches []func(*primitive.Map) bool
		dataLock      sync.RWMutex
		streamLock    sync.RWMutex
	}

	fullEvent struct {
		database.Event
		Document *primitive.Map
	}
)

var _ database.Collection = &Collection{}

var (
	ErrCodePKNotFound   = "primary key is not found"
	ErrCodePKDuplicated = "primary key is duplicated"

	ErrPKNotFound   = errors.New(ErrCodePKNotFound)
	ErrPKDuplicated = errors.New(ErrCodePKDuplicated)
)

func NewCollection(name string) *Collection {
	return &Collection{
		name:       name,
		data:       pool.GetMap(),
		indexView:  NewIndexView(),
		dataLock:   sync.RWMutex{},
		streamLock: sync.RWMutex{},
	}
}

func (coll *Collection) Name() string {
	coll.dataLock.RLock()
	defer coll.dataLock.RUnlock()

	return coll.name
}

func (coll *Collection) Indexes() database.IndexView {
	coll.dataLock.RLock()
	defer coll.dataLock.RUnlock()

	return coll.indexView
}

func (coll *Collection) Watch(ctx context.Context, filter *database.Filter) (database.Stream, error) {
	coll.streamLock.Lock()
	defer coll.streamLock.Unlock()

	stream := NewStream()
	coll.streams = append(coll.streams, stream)
	coll.streamMatches = append(coll.streamMatches, ParseFilter(filter))

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

func (coll *Collection) InsertOne(ctx context.Context, doc *primitive.Map) (primitive.Object, error) {
	if id, err := coll.insertOne(ctx, doc); err != nil {
		return nil, err
	} else {
		coll.emit(fullEvent{
			Event: database.Event{
				OP:         database.EventInsert,
				DocumentID: id,
			},
			Document: doc,
		})
		return id, nil
	}
}

func (coll *Collection) InsertMany(ctx context.Context, docs []*primitive.Map) ([]primitive.Object, error) {
	if ids, err := coll.insertMany(ctx, docs); err != nil {
		return nil, err
	} else {
		for i, doc := range docs {
			coll.emit(fullEvent{
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
		upsert = util.UnPtr(opt.Upsert)
	}

	old, err := coll.findOne(ctx, filter)
	if err != nil {
		return false, err
	}
	if old == nil && !upsert {
		return false, nil
	}

	var id primitive.Object
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

	coll.emit(fullEvent{
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
		upsert = util.UnPtr(opt.Upsert)
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
	if ids, err := coll.insertMany(ctx, docs); err != nil {
		_, _ = coll.insertMany(ctx, old)
		return 0, err
	} else {
		for i, doc := range docs {
			coll.emit(fullEvent{
				Event: database.Event{
					OP:         database.EventUpdate,
					DocumentID: ids[i],
				},
				Document: doc,
			})
		}
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
				coll.emit(fullEvent{
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
				coll.emit(fullEvent{
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
	data, err := func() (*sync.Map, error) {
		coll.dataLock.Lock()
		defer coll.dataLock.Unlock()

		data := coll.data
		coll.data = pool.GetMap()

		if err := coll.indexView.deleteAll(ctx); err != nil {
			return nil, err
		}

		return data, nil
	}()
	if err != nil {
		return err
	}

	data.Range(func(_, val any) bool {
		doc := val.(*primitive.Map)
		if id, ok := doc.Get(keyID); ok {
			coll.emit(fullEvent{
				Event: database.Event{
					OP:         database.EventDelete,
					DocumentID: id,
				},
				Document: doc,
			})
		}
		return true
	})

	coll.streamLock.Lock()
	defer coll.streamLock.Unlock()

	for _, s := range coll.streams {
		if err := s.Close(); err != nil {
			return err
		}
	}
	coll.streams = nil

	return nil
}

func (coll *Collection) insertOne(ctx context.Context, doc *primitive.Map) (primitive.Object, error) {
	if ids, err := coll.insertMany(ctx, []*primitive.Map{doc}); err != nil {
		return nil, err
	} else {
		return ids[0], nil
	}
}

func (coll *Collection) insertMany(ctx context.Context, docs []*primitive.Map) ([]primitive.Object, error) {
	coll.dataLock.Lock()
	defer coll.dataLock.Unlock()

	ids := make([]primitive.Object, len(docs))
	for i, doc := range docs {
		if id, ok := doc.Get(keyID); !ok {
			return nil, errors.Wrap(errors.WithStack(ErrPKNotFound), database.ErrCodeWrite)
		} else if hash, err := util.Hash(id); err != nil {
			return nil, errors.Wrap(err, database.ErrCodeWrite)
		} else if _, ok := coll.data.Load(hash); ok {
			return nil, errors.Wrap(errors.WithStack(ErrPKDuplicated), database.ErrCodeWrite)
		} else {
			ids[i] = id
		}
	}

	if err := coll.indexView.insertMany(ctx, docs); err != nil {
		return nil, errors.Wrap(err, database.ErrCodeWrite)
	}
	for i, doc := range docs {
		if hash, err := util.Hash(ids[i].Interface()); err != nil {
			return nil, errors.Wrap(err, database.ErrCodeWrite)
		} else {
			coll.data.Store(hash, doc)
		}
	}

	return ids, nil
}

func (coll *Collection) findOne(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) (*primitive.Map, error) {
	opt := database.MergeFindOptions(append(opts, util.Ptr(database.FindOptions{Limit: util.Ptr(1)})))

	if docs, err := coll.findMany(ctx, filter, opt); err != nil {
		return nil, err
	} else if len(docs) > 0 {
		return docs[0], nil
	} else {
		return nil, nil
	}
}

func (coll *Collection) findMany(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) ([]*primitive.Map, error) {
	coll.dataLock.RLock()
	defer coll.dataLock.RUnlock()

	opt := database.MergeFindOptions(opts)

	limit := -1
	if opt != nil && opt.Limit != nil {
		limit = util.UnPtr(opt.Limit)
	}
	skip := 0
	if opt != nil && opt.Skip != nil {
		skip = util.UnPtr(opt.Skip)
	}
	var sorts []database.Sort
	if opt != nil && opt.Sorts != nil {
		sorts = opt.Sorts
	}

	match := ParseFilter(filter)

	scanSize := limit
	if skip > 0 || len(sorts) > 0 {
		scanSize = -1
	}

	scan := map[uint64]*primitive.Map{}
	if examples, ok := FilterToExample(filter); ok {
		if ids, err := coll.indexView.findMany(ctx, examples); err == nil {
			for _, id := range ids {
				if scanSize == len(scan) {
					break
				} else if hash, err := util.Hash(id.Interface()); err != nil {
					return nil, errors.Wrap(err, database.ErrCodeWrite)
				} else if doc, ok := coll.data.Load(hash); ok && match(doc.(*primitive.Map)) {
					scan[hash] = doc.(*primitive.Map)
				}
			}
		}
	}
	if scanSize != len(scan) {
		coll.data.Range(func(key, value any) bool {
			if scanSize == len(scan) {
				return false
			}

			if match(value.(*primitive.Map)) {
				scan[key.(uint64)] = value.(*primitive.Map)
			}
			return true
		})
	}

	if skip >= len(scan) {
		return nil, nil
	}

	var docs []*primitive.Map
	for _, doc := range scan {
		docs = append(docs, doc)
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
	coll.dataLock.Lock()
	defer coll.dataLock.Unlock()

	ids := make([]primitive.Object, 0, len(docs))
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
		if hash, err := util.Hash(id.Interface()); err != nil {
			return nil, errors.Wrap(err, database.ErrCodeWrite)
		} else {
			coll.data.Delete(hash)
		}
	}

	return deletes, nil
}

func (coll *Collection) unwatch(stream database.Stream) {
	coll.streamLock.Lock()
	defer coll.streamLock.Unlock()

	for i, s := range coll.streams {
		if s == stream {
			coll.streams = append(coll.streams[:i], coll.streams[i+1:]...)
			coll.streamMatches = append(coll.streamMatches[:i], coll.streamMatches[i+1:]...)
			return
		}
	}
}

func (coll *Collection) emit(event fullEvent) {
	coll.streamLock.RLock()
	defer coll.streamLock.RUnlock()

	for i, s := range coll.streams {
		if coll.streamMatches[i](event.Document) {
			s.Emit(event.Event)
		}
	}
}
