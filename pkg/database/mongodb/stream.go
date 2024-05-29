package mongodb

import (
	"context"
	"sync"

	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/object"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Stream struct {
	internal *mongo.ChangeStream
	channel  chan database.Event
	done     chan struct{}
	mu       sync.Mutex
}

func newStream(ctx context.Context, stream *mongo.ChangeStream) *Stream {
	s := &Stream{
		internal: stream,
		channel:  make(chan database.Event),
		done:     make(chan struct{}),
	}

	go func() {
		defer func() { _ = s.internal.Close(ctx) }()
		defer close(s.channel)

		ctx, cancel := context.WithCancel(ctx)
		go func() {
			defer cancel()
			<-s.done
		}()

		for {
			if !s.internal.Next(ctx) {
				return
			}
			var data bson.M
			if err := stream.Decode(&data); err != nil {
				return
			}

			var id object.Object
			if documentKey, ok := data["documentKey"]; ok {
				if documentKey, ok := documentKey.(bson.M); ok {
					if err := bsonToPrimitive(documentKey["_id"], &id); err != nil {
						continue
					}
				} else {
					continue
				}
			}

			e := database.Event{
				DocumentID: id,
			}
			switch data["operationType"] {
			case "insert":
				e.OP = database.EventInsert
			case "update":
				e.OP = database.EventUpdate
			case "delete":
				e.OP = database.EventDelete
			}

			select {
			case <-s.done:
				return
			case s.channel <- e:
			}
		}
	}()

	return s
}

func (s *Stream) Next() <-chan database.Event {
	return s.channel
}

func (s *Stream) Done() <-chan struct{} {
	return s.done
}

func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return nil
	default:
	}

	close(s.done)

	return nil
}
