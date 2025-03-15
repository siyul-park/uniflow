package store

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Stream struct to hold the change stream
type stream struct {
	changeStream *mongo.ChangeStream
	mu           sync.Mutex
}

var _ store.Stream = (*stream)(nil)

// Next moves the change stream to the next event
func (s *stream) Next(ctx context.Context) bool {
	for {
		s.mu.Lock()
		if s.changeStream.TryNext(ctx) {
			s.mu.Unlock()
			return true
		}
		if err := s.changeStream.Err(); err != nil {
			s.mu.Unlock()
			return false
		}
		s.mu.Unlock()
	}
}

// Decode takes the next MongoDB change stream event and converts it into an Event struct
func (s *stream) Decode(val any) error {
	var raw bson.M
	if err := s.changeStream.Decode(&raw); err != nil {
		return err
	}

	v, err := types.Cast[types.Map](fromBSON(raw))
	if err != nil {
		return err
	}

	event := &store.Event{}
	if err := types.Unmarshal(v.Get(types.NewString("id")), &event.ID); err != nil {
		return err
	}
	if err := types.Unmarshal(v.Get(types.NewString("operationType")), &event.OP); err != nil {
		return err
	}

	value, err := types.Marshal(event)
	if err != nil {
		return err
	}
	return types.Unmarshal(value, val)
}

// Close stops the change stream
func (s *stream) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.changeStream.Close(ctx)
}
