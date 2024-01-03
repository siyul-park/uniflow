package memdb

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type Collection struct {
	name    string
	segment *Segment
	streams []*Stream
	matches []func(*primitive.Map) bool
	mu      sync.RWMutex
}

type internalEvent struct {
	op       database.EventOP
	document *primitive.Map
}

var _ database.Collection = &Collection{}

var (
	ErrPKNotFound   = errors.New("primary key is not found")
	ErrPKDuplicated = errors.New("primary key is duplicated")
)

func NewCollection(name string) *Collection {
	return &Collection{
		name:    name,
		segment: newSegment(),
		mu:      sync.RWMutex{},
	}
}

func (c *Collection) Name() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.name
}

func (c *Collection) Indexes() database.IndexView {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return newIndexView(c.segment)
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
	ids, err := c.segment.Insert([]*primitive.Map{doc})
	if err != nil {
		return nil, err
	}

	c.emit(internalEvent{op: database.EventInsert, document: doc})

	return ids[0], nil
}

func (c *Collection) InsertMany(_ context.Context, docs []*primitive.Map) ([]primitive.Value, error) {
	ids, err := c.segment.Insert(docs)
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		c.emit(internalEvent{op: database.EventInsert, document: doc})
	}

	return ids, nil
}

func (c *Collection) UpdateOne(_ context.Context, filter *database.Filter, patch *primitive.Map, opts ...*database.UpdateOptions) (bool, error) {
	opt := database.MergeUpdateOptions(opts)

	upsert := false
	if opt != nil && opt.Upsert != nil {
		upsert = lo.FromPtr(opt.Upsert)
	}

	origins, err := c.segment.Find(filter, &database.FindOptions{Limit: lo.ToPtr[int](1)})
	if err != nil {
		return false, err
	}

	if len(origins) == 0 && !upsert {
		return false, nil
	}

	var id primitive.Value
	if len(origins) > 0 {
		id = origins[0].GetOr(keyID, nil)
	}
	if id == nil {
		id = patch.GetOr(keyID, extractIDByFilter(filter))
	}
	if id == nil {
		return false, errors.Wrap(errors.WithStack(ErrPKNotFound), database.ErrCodeWrite)
	}

	if origins != nil {
		if _, err := c.segment.Delete(origins); err != nil {
			return false, err
		}
	}

	if _, ok := patch.Get(keyID); !ok {
		patch = patch.Set(keyID, id)
	}

	if _, err := c.segment.Insert([]*primitive.Map{patch}); err != nil {
		_, _ = c.segment.Insert(origins)
		return false, err
	}

	c.emit(internalEvent{op: database.EventUpdate, document: patch})

	return true, nil
}

func (c *Collection) UpdateMany(_ context.Context, filter *database.Filter, patch *primitive.Map, opts ...*database.UpdateOptions) (int, error) {
	opt := database.MergeUpdateOptions(opts)
	upsert := false
	if opt != nil && opt.Upsert != nil {
		upsert = lo.FromPtr(opt.Upsert)
	}

	origins, err := c.segment.Find(filter)
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
		if _, err := c.segment.Insert([]*primitive.Map{patch}); err != nil {
			return 0, err
		}
		return 1, nil
	}

	if _, err := c.segment.Delete(origins); err != nil {
		return 0, err
	}

	patches := make([]*primitive.Map, len(origins))
	for i, origin := range origins {
		patches[i] = patch.Set(keyID, origin.GetOr(keyID, nil))
	}

	if _, err := c.segment.Insert(patches); err != nil {
		_, _ = c.segment.Insert(origins)
		return 0, err
	}

	for _, patch := range patches {
		c.emit(internalEvent{op: database.EventUpdate, document: patch})
	}

	return len(patches), nil
}

func (c *Collection) DeleteOne(_ context.Context, filter *database.Filter) (bool, error) {
	if docs, err := c.segment.Find(filter, &database.FindOptions{Limit: lo.ToPtr[int](1)}); err != nil {
		return false, err
	} else if origins, err := c.segment.Delete(docs); err != nil {
		return false, err
	} else if len(origins) == 0 {
		return false, nil
	} else {
		c.emit(internalEvent{op: database.EventDelete, document: origins[0]})
		return true, nil
	}
}

func (c *Collection) DeleteMany(_ context.Context, filter *database.Filter) (int, error) {
	if docs, err := c.segment.Find(filter); err != nil {
		return 0, err
	} else if origins, err := c.segment.Delete(docs); err != nil {
		return 0, err
	} else {
		for _, doc := range origins {
			c.emit(internalEvent{op: database.EventDelete, document: doc})
		}
		return len(origins), nil
	}
}

func (c *Collection) FindOne(_ context.Context, filter *database.Filter, opts ...*database.FindOptions) (*primitive.Map, error) {
	docs, err := c.segment.Find(filter, &database.FindOptions{Limit: lo.ToPtr[int](1)})
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	return docs[0], nil
}

func (c *Collection) FindMany(_ context.Context, filter *database.Filter, opts ...*database.FindOptions) ([]*primitive.Map, error) {
	return c.segment.Find(filter, opts...)
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
