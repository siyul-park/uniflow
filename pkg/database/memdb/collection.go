package memdb

import (
	"context"
	"sort"
	"sync"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/types"
)

// Collection represents a collection of documents in a database.
type Collection struct {
	name      string
	section   *Section
	indexView *IndexView
	streams   []*Stream
	matches   []func(types.Map) bool
	mu        sync.RWMutex
}

type event struct {
	op       database.EventOP
	document types.Map
}

var _ database.Collection = (*Collection)(nil)

// NewCollection creates a new collection with the given name.
func NewCollection(name string) *Collection {
	segment := newSection()

	return &Collection{
		name:      name,
		section:   segment,
		indexView: newIndexView(segment),
		mu:        sync.RWMutex{},
	}
}

// Name returns the name of the collection.
func (c *Collection) Name() string {
	return c.name
}

// Indexes returns the index view of the collection.
func (c *Collection) Indexes() database.IndexView {
	return c.indexView
}

// Watch sets up a stream to watch for changes that match the given filter.
func (c *Collection) Watch(ctx context.Context, filter *database.Filter) (database.Stream, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	stream := newStream()

	// Add the stream and its filter match function to the collection.
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

// InsertOne inserts a single document into the collection.
func (c *Collection) InsertOne(_ context.Context, doc types.Map) (types.Value, error) {
	id, err := c.section.Set(doc)
	if err != nil {
		return nil, errors.Wrap(err, database.ErrCodeWrite)
	}

	c.emit(event{op: database.EventInsert, document: doc})
	return id, nil
}

// InsertMany inserts multiple documents into the collection.
func (c *Collection) InsertMany(_ context.Context, docs []types.Map) ([]types.Value, error) {
	ids := make([]types.Value, len(docs))
	for i, doc := range docs {
		id, err := c.section.Set(doc)
		if err != nil {
			for ; i >= 0; i-- {
				c.section.Delete(docs[i])
			}
			return nil, errors.Wrap(err, database.ErrCodeWrite)
		}
		ids[i] = id
	}

	for _, doc := range docs {
		c.emit(event{op: database.EventInsert, document: doc})
	}
	return ids, nil
}

// UpdateOne updates a single document in the collection that matches the filter.
func (c *Collection) UpdateOne(ctx context.Context, filter *database.Filter, patch types.Map, opts ...*database.UpdateOptions) (bool, error) {
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

	var id types.Value
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
		c.section.Delete(origin)
	}

	if _, ok := patch.Get(keyID); !ok {
		patch = patch.Set(keyID, id)
	}

	if _, err := c.section.Set(patch); err != nil {
		_, _ = c.section.Set(origin)
		return false, errors.Wrap(err, database.ErrCodeWrite)
	}

	c.emit(event{op: database.EventUpdate, document: patch})
	return true, nil
}

// UpdateMany updates multiple documents in the collection that match the filter.
func (c *Collection) UpdateMany(ctx context.Context, filter *database.Filter, patch types.Map, opts ...*database.UpdateOptions) (int, error) {
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
		if _, err := c.section.Set(patch); err != nil {
			return 0, errors.Wrap(err, database.ErrCodeWrite)
		}
		return 1, nil
	}

	for _, origin := range origins {
		c.section.Delete(origin)
	}

	patches := make([]types.Map, len(origins))
	for i, origin := range origins {
		patches[i] = patch.Set(keyID, origin.GetOr(keyID, nil))
	}

	for i, patch := range patches {
		if _, err := c.section.Set(patch); err != nil {
			for ; i >= 0; i-- {
				_, _ = c.section.Set(origins[i])
			}
			return 0, errors.Wrap(err, database.ErrCodeWrite)
		}
	}

	for _, patch := range patches {
		c.emit(event{op: database.EventUpdate, document: patch})
	}
	return len(patches), nil
}

// DeleteOne deletes a single document in the collection that matches the filter.
func (c *Collection) DeleteOne(ctx context.Context, filter *database.Filter) (bool, error) {
	if doc, err := c.FindOne(ctx, filter); err != nil || doc == nil {
		return false, err
	} else {
		c.section.Delete(doc)
		c.emit(event{op: database.EventDelete, document: doc})
		return true, nil
	}
}

// DeleteMany deletes multiple documents in the collection that match the filter.
func (c *Collection) DeleteMany(ctx context.Context, filter *database.Filter) (int, error) {
	if docs, err := c.FindMany(ctx, filter); err != nil {
		return 0, err
	} else {
		for _, doc := range docs {
			c.section.Delete(doc)
			c.emit(event{op: database.EventDelete, document: doc})
		}
		return len(docs), nil
	}
}

// FindOne finds a single document in the collection that matches the filter.
func (c *Collection) FindOne(ctx context.Context, filter *database.Filter, opts ...*database.FindOptions) (types.Map, error) {
	docs, err := c.FindMany(ctx, filter, &database.FindOptions{Limit: lo.ToPtr[int](1)})
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	return docs[0], nil
}

// FindMany finds multiple documents in the collection that match the filter.
func (c *Collection) FindMany(_ context.Context, filter *database.Filter, opts ...*database.FindOptions) ([]types.Map, error) {
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
	fullScan := true
	var plan *executionPlan
	for _, constraint := range c.section.Constraints() {
		if cur := newExecutionPlan(constraint.Keys, filter); cur != nil && (plan == nil || plan.Cost() > cur.Cost()) {
			cur.key = constraint.Name
			fullScan = constraint.Partial != nil
			plan = cur
		}
	}

	scan := newNodes()
	defer deleteNodes(scan)

	appends := func(doc types.Map) bool {
		if match == nil || match(doc) {
			scan.Set(node{key: doc.GetOr(keyID, nil), value: doc})
		}
		return len(sorts) > 0 || limit < 0 || scan.Len() < limit+skip
	}

	if plan != nil {
		sector, ok := c.section.Scan(plan.key, plan.min, plan.max)
		plan = plan.next

		for ok && plan != nil {
			sector, ok = sector.Scan(plan.key, plan.min, plan.max)
			plan = plan.next
		}

		if ok {
			sector.Range(appends)
		} else {
			fullScan = true
		}
	}

	if fullScan {
		c.section.Range(appends)
	}

	var docs []types.Map
	scan.Scan(func(n node) bool {
		docs = append(docs, n.value)
		return true
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

// Drop drops the collection and closes all active streams.
func (c *Collection) Drop(_ context.Context) error {
	data := c.section.Drop()
	for _, doc := range data {
		c.emit(event{op: database.EventDelete, document: doc})
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, s := range c.streams {
		if err := s.Close(); err != nil {
			return err
		}
	}
	c.streams = nil

	return nil
}

// unwatch removes the stream from the collection's active streams.
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

// emit emits an internal event to all matching streams.
func (c *Collection) emit(evt event) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for i, s := range c.streams {
		match := c.matches[i]
		if match == nil || match(evt.document) {
			s.Emit(database.Event{
				OP:         evt.op,
				DocumentID: evt.document.GetOr(keyID, nil),
			})
		}
	}
}
