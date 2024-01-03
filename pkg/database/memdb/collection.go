package memdb

import (
	"context"
	"sort"
	"sync"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type Collection struct {
	name      string
	segment   *Segment
	indexView *IndexView
	streams   []*Stream
	matches   []func(*primitive.Map) bool
	mu        sync.RWMutex
}

type internalEvent struct {
	op       database.EventOP
	document *primitive.Map
}

var _ database.Collection = &Collection{}

func NewCollection(name string) *Collection {
	segment := newSegment()

	return &Collection{
		name:      name,
		segment:   segment,
		indexView: newIndexView(segment),
		mu:        sync.RWMutex{},
	}
}

func (c *Collection) Name() string {
	return c.name
}

func (c *Collection) Indexes() database.IndexView {
	return c.indexView
}

func (c *Collection) Watch(ctx context.Context, filter *database.Filter) (database.Stream, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	stream := NewStream()

	c.streams = append(c.streams, stream)
	c.matches = append(c.matches, parseFilter(filter))

	go func() {
		select {
		case <-stream.Done():
			c.unwatch(stream)
		case <-ctx.Done():
			_ = stream.Close()
			c.unwatch(stream)
		}
	}()

	return stream, nil
}

func (c *Collection) InsertOne(_ context.Context, doc *primitive.Map) (primitive.Value, error) {
	ids, err := c.segment.Set([]*primitive.Map{doc})
	if err != nil {
		return nil, errors.Wrap(err, database.ErrCodeWrite)
	}

	c.emit(internalEvent{op: database.EventInsert, document: doc})

	return ids[0], nil
}

func (c *Collection) InsertMany(_ context.Context, docs []*primitive.Map) ([]primitive.Value, error) {
	ids, err := c.segment.Set(docs)
	if err != nil {
		return nil, errors.Wrap(err, database.ErrCodeWrite)
	}

	for _, doc := range docs {
		c.emit(internalEvent{op: database.EventInsert, document: doc})
	}

	return ids, nil
}

func (c *Collection) UpdateOne(ctx context.Context, filter *database.Filter, patch *primitive.Map, opts ...*database.UpdateOptions) (bool, error) {
	opt := database.MergeUpdateOptions(opts)

	upsert := false
	if opt != nil && opt.Upsert != nil {
		upsert = lo.FromPtr(opt.Upsert)
	}

	origin, err := c.FindOne(ctx, filter)
	if err != nil {
		return false, err
	}

	if origin == nil && !upsert {
		return false, nil
	}

	var id primitive.Value
	if origin != nil {
		id = origin.GetOr(keyID, nil)
	}
	if id == nil {
		id = patch.GetOr(keyID, extractIDByFilter(filter))
	}
	if id == nil {
		return false, errors.Wrap(errors.WithStack(ErrPKNotFound), database.ErrCodeWrite)
	}

	if origin != nil {
		_ = c.segment.Delete([]*primitive.Map{origin})
	}

	if _, ok := patch.Get(keyID); !ok {
		patch = patch.Set(keyID, id)
	}

	if _, err := c.segment.Set([]*primitive.Map{patch}); err != nil {
		_, _ = c.segment.Set([]*primitive.Map{origin})
		return false, errors.Wrap(err, database.ErrCodeWrite)
	}

	c.emit(internalEvent{op: database.EventUpdate, document: patch})

	return true, nil
}

func (c *Collection) UpdateMany(ctx context.Context, filter *database.Filter, patch *primitive.Map, opts ...*database.UpdateOptions) (int, error) {
	opt := database.MergeUpdateOptions(opts)
	upsert := false
	if opt != nil && opt.Upsert != nil {
		upsert = lo.FromPtr(opt.Upsert)
	}

	origins, err := c.FindMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	if len(origins) == 0 {
		if !upsert {
			return 0, nil
		}

		id := patch.GetOr(keyID, extractIDByFilter(filter))
		if id == nil {
			return 0, errors.Wrap(errors.WithStack(ErrPKNotFound), database.ErrCodeWrite)
		}

		if _, ok := patch.Get(keyID); !ok {
			patch = patch.Set(keyID, id)
		}
		if _, err := c.segment.Set([]*primitive.Map{patch}); err != nil {
			return 0, errors.Wrap(err, database.ErrCodeWrite)
		}
		return 1, nil
	}

	_ = c.segment.Delete(origins)

	patches := make([]*primitive.Map, len(origins))
	for i, origin := range origins {
		patches[i] = patch.Set(keyID, origin.GetOr(keyID, nil))
	}

	if _, err := c.segment.Set(patches); err != nil {
		_, _ = c.segment.Set(origins)
		return 0, errors.Wrap(err, database.ErrCodeWrite)
	}

	for _, patch := range patches {
		c.emit(internalEvent{op: database.EventUpdate, document: patch})
	}

	return len(patches), nil
}

func (c *Collection) DeleteOne(ctx context.Context, filter *database.Filter) (bool, error) {
	if doc, err := c.FindOne(ctx, filter); err != nil || doc == nil {
		return false, err
	} else if origins := c.segment.Delete([]*primitive.Map{doc}); len(origins) == 0 {
		return false, nil
	} else {
		c.emit(internalEvent{op: database.EventDelete, document: origins[0]})
		return true, nil
	}
}

func (c *Collection) DeleteMany(ctx context.Context, filter *database.Filter) (int, error) {
	if docs, err := c.FindMany(ctx, filter); err != nil {
		return 0, err
	} else {
		origins := c.segment.Delete(docs)
		for _, doc := range origins {
			c.emit(internalEvent{op: database.EventDelete, document: doc})
		}
		return len(origins), nil
	}
}

func (c *Collection) FindOne(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) (*primitive.Map, error) {
	docs, err := c.FindMany(ctx, filter, &database.FindOptions{Limit: lo.ToPtr[int](1)})
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	return docs[0], nil
}

func (c *Collection) FindMany(_ context.Context, filter *database.Filter, opts ...*database.FindOptions) ([]*primitive.Map, error) {
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

	match := parseFilter(filter)

	var docs []*primitive.Map
	c.segment.Range(func(doc *primitive.Map) bool {
		if match(doc) {
			docs = append(docs, doc)
		}
		return len(sorts) > 0 || limit < 0 || len(docs) < limit+skip
	})

	if skip >= len(docs) {
		return nil, nil
	}
	if len(sorts) > 0 {
		compare := parseSorts(sorts)
		sort.Slice(docs, func(i, j int) bool {
			return compare(docs[i], docs[j])
		})
	}
	docs = docs[skip:]
	if limit >= 0 && len(docs) > limit {
		docs = docs[:limit]
	}

	return docs, nil
}

func (coll *Collection) Drop(_ context.Context) error {
	data := coll.segment.Drop()
	for _, doc := range data {
		coll.emit(internalEvent{op: database.EventDelete, document: doc})
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

func (c *Collection) unwatch(stream database.Stream) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, s := range c.streams {
		if s == stream {
			c.streams = append(c.streams[:i], c.streams[i+1:]...)
			c.matches = append(c.matches[:i], c.matches[i+1:]...)
			return
		}
	}
}

func (c *Collection) emit(evt internalEvent) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for i, s := range c.streams {
		if c.matches[i](evt.document) {
			s.Emit(database.Event{
				OP:         evt.op,
				DocumentID: evt.document.GetOr(keyID, nil),
			})
		}
	}
}
