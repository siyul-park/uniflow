package mongo

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/spec"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Store manages storage and retrieval of Spec objects in a MongoDB collection.
type Store struct {
	collection *mongo.Collection
}

// Stream represents a MongoDB change stream for tracking Spec changes.
type Stream struct {
	stream *mongo.ChangeStream
	ctx    context.Context
	cancel context.CancelFunc
	out    chan spec.Event
}

type changeDocument struct {
	OperationType string `bson:"operationType"`
	DocumentKey   struct {
		ID uuid.UUID `bson:"_id"`
	} `bson:"documentKey"`
}

var _ spec.Store = (*Store)(nil)
var _ spec.Stream = (*Stream)(nil)

// NewStore creates a new Store with the specified MongoDB collection.
func NewStore(collection *mongo.Collection) *Store {
	return &Store{collection: collection}
}

// Index ensures the collection has the required indexes and updates them if necessary.
func (s *Store) Index(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: spec.KeyNamespace, Value: 1},
				{Key: spec.KeyName, Value: 1},
			},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{
				spec.KeyName: bson.M{"$exists": true, "$ne": ""},
			}),
		},
		{
			Keys: bson.M{spec.KeyKind: 1},
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Watch returns a Stream that monitors changes matching the specified filter.
func (s *Store) Watch(ctx context.Context, specs ...spec.Spec) (spec.Stream, error) {
	filter := s.filter(specs...)

	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	stream, err := s.collection.Watch(ctx, mongo.Pipeline{bson.D{{Key: "$match", Value: filter}}}, opts)
	if err != nil {
		return nil, err
	}

	return newStream(stream), nil
}

// Load retrieves Specs from the store that match the given criteria.
func (s *Store) Load(ctx context.Context, specs ...spec.Spec) ([]spec.Spec, error) {
	filter := s.filter(specs...)
	limit := int64(s.limit(specs...))

	cursor, err := s.collection.Find(ctx, filter, &options.FindOptions{
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []spec.Spec
	for cursor.Next(ctx) {
		var unstructured spec.Unstructured
		if err := cursor.Decode(&unstructured); err != nil {
			return nil, err
		}
		result = append(result, &unstructured)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// Store saves the given Specs into the database.
func (s *Store) Store(ctx context.Context, specs ...spec.Spec) (int, error) {
	var docs []interface{}
	for _, v := range specs {
		if v.GetID() == uuid.Nil {
			v.SetID(uuid.Must(uuid.NewV7()))
		}
		if v.GetNamespace() == "" {
			v.SetNamespace(spec.DefaultNamespace)
		}

		docs = append(docs, v)
	}

	res, err := s.collection.InsertMany(ctx, docs)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return 0, errors.WithMessage(spec.ErrDuplicatedKey, err.Error())
		}
		return 0, err
	}
	return len(res.InsertedIDs), nil
}

// Swap updates existing Specs in the database with the provided data.
func (s *Store) Swap(ctx context.Context, specs ...spec.Spec) (int, error) {
	count := 0
	for _, v := range specs {
		if v.GetID() == uuid.Nil {
			v.SetID(uuid.Must(uuid.NewV7()))
		}
		if v.GetNamespace() == "" {
			v.SetNamespace(spec.DefaultNamespace)
		}

		filter := bson.M{"_id": v.GetID()}
		update := bson.M{"$set": v}

		res, err := s.collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return 0, err
		}
		count += int(res.MatchedCount)
	}
	return count, nil
}

// Delete removes Specs from the store based on the provided criteria.
func (s *Store) Delete(ctx context.Context, specs ...spec.Spec) (int, error) {
	filter := s.filter(specs...)
	res, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(res.DeletedCount), nil
}

func (s *Store) filter(specs ...spec.Spec) bson.M {
	var orFilters []bson.M
	for _, v := range specs {
		andFilters := bson.M{}
		if v.GetID() != uuid.Nil {
			andFilters["_id"] = v.GetID()
		}
		if v.GetNamespace() != "" {
			andFilters[spec.KeyNamespace] = v.GetNamespace()
		}
		if v.GetName() != "" {
			andFilters[spec.KeyName] = v.GetName()
		}
		if len(andFilters) > 0 {
			orFilters = append(orFilters, andFilters)
		}
	}

	switch len(orFilters) {
	case 0:
		return bson.M{}
	case 1:
		return orFilters[0]
	default:
		return bson.M{"$or": orFilters}
	}
}

func (s *Store) limit(specs ...spec.Spec) int {
	limit := 0
	for _, v := range specs {
		if v.GetID() != uuid.Nil || v.GetName() != "" {
			limit += 1
		} else if v.GetNamespace() != "" {
			return 0
		}
	}
	return limit
}

// newStream creates and returns a new Stream.
func newStream(stream *mongo.ChangeStream) *Stream {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Stream{
		stream: stream,
		ctx:    ctx,
		cancel: cancel,
		out:    make(chan spec.Event),
	}

	go func() {
		defer close(s.out)

		for s.stream.Next(s.ctx) {
			var doc changeDocument
			if err := s.stream.Decode(&doc); err != nil {
				continue
			}

			var op spec.EventOP
			switch doc.OperationType {
			case "insert":
				op = spec.EventStore
			case "update":
				op = spec.EventSwap
			case "delete":
				op = spec.EventDelete
			default:
				continue
			}

			event := spec.Event{
				OP: op,
				ID: doc.DocumentKey.ID,
			}

			select {
			case <-ctx.Done():
				return
			case s.out <- event:
			}
		}
	}()

	return s
}

// Next returns a channel for receiving events from the stream.
func (s *Stream) Next() <-chan spec.Event {
	return s.out
}

// Done returns a channel that is closed when the stream is closed.
func (s *Stream) Done() <-chan struct{} {
	return s.ctx.Done()
}

// Close closes the stream and releases any resources.
func (s *Stream) Close() error {
	s.cancel()
	return nil
}
