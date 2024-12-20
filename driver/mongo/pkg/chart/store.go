package chart

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	_ "github.com/siyul-park/uniflow/driver/mongo/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/secret"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	out    chan resource.Event
}

type changeDocument struct {
	OperationType string `bson:"operationType"`
	DocumentKey   struct {
		ID uuid.UUID `bson:"_id"`
	} `bson:"documentKey"`
}

var _ chart.Store = (*Store)(nil)
var _ chart.Stream = (*Stream)(nil)

// NewStore creates a new Store with the specified MongoDB collection.
func NewStore(collection *mongo.Collection) *Store {
	return &Store{collection: collection}
}

// Index ensures the collection has the required indexes and updates them if necessary.
func (s *Store) Index(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: chart.KeyNamespace, Value: 1},
				{Key: chart.KeyName, Value: 1},
			},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{
				chart.KeyName: bson.M{"$exists": true},
			}),
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Watch returns a Stream that monitors changes matching the specified filter.
func (s *Store) Watch(ctx context.Context, charts ...*chart.Chart) (secret.Stream, error) {
	filter := s.filter(charts...)

	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	changeStream, err := s.collection.Watch(ctx, mongo.Pipeline{bson.D{{Key: "$match", Value: filter}}}, opts)
	if err != nil {
		return nil, err
	}

	stream := newStream(changeStream)

	go func() {
		select {
		case <-ctx.Done():
			stream.Close()
		case <-stream.Done():
		}
	}()

	return stream, nil
}

// Load retrieves Specs from the store that match the given criteria.
func (s *Store) Load(ctx context.Context, charts ...*chart.Chart) ([]*chart.Chart, error) {
	filter := s.filter(charts...)
	limit := int64(s.limit(charts...))

	cursor, err := s.collection.Find(ctx, filter, options.Find().SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []*chart.Chart
	for cursor.Next(ctx) {
		chrt := &chart.Chart{}
		if err := cursor.Decode(&chrt); err != nil {
			return nil, err
		}
		result = append(result, chrt)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// Store saves the given Specs into the database.
func (s *Store) Store(ctx context.Context, charts ...*chart.Chart) (int, error) {
	var docs []any
	for _, chrt := range charts {
		if chrt.GetID() == uuid.Nil {
			chrt.SetID(uuid.Must(uuid.NewV7()))
		}
		if chrt.GetNamespace() == "" {
			chrt.SetNamespace(resource.DefaultNamespace)
		}

		docs = append(docs, chrt)
	}

	res, err := s.collection.InsertMany(ctx, docs)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return 0, errors.WithMessage(resource.ErrDuplicatedKey, err.Error())
		}
		return 0, err
	}
	return len(res.InsertedIDs), nil
}

// Swap updates existing Specs in the database with the provided data.
func (s *Store) Swap(ctx context.Context, charts ...*chart.Chart) (int, error) {
	ids := make([]uuid.UUID, len(charts))
	for i, chrt := range charts {
		if chrt.GetID() == uuid.Nil {
			chrt.SetID(uuid.Must(uuid.NewV7()))
		}
		ids[i] = chrt.GetID()
	}

	filter := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	ok := make(map[uuid.UUID]*chart.Chart)
	for cursor.Next(ctx) {
		chrt := &chart.Chart{}
		if err := cursor.Decode(&chrt); err != nil {
			return 0, err
		}
		ok[chrt.GetID()] = chrt
	}

	count := 0
	for _, chrt := range charts {
		exist, ok := ok[chrt.GetID()]
		if !ok {
			continue
		}

		chrt.SetNamespace(exist.GetNamespace())

		filter := bson.M{"_id": chrt.GetID()}
		update := bson.M{"$set": chrt}

		res, err := s.collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return 0, err
		}
		count += int(res.MatchedCount)
	}
	return count, nil
}

// Delete removes Specs from the store based on the provided criteria.
func (s *Store) Delete(ctx context.Context, charts ...*chart.Chart) (int, error) {
	filter := s.filter(charts...)
	res, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int(res.DeletedCount), nil
}

func (s *Store) filter(charts ...*chart.Chart) bson.M {
	var orFilters []bson.M
	for _, v := range charts {
		andFilters := bson.M{}
		if v.GetID() != uuid.Nil {
			andFilters["_id"] = v.GetID()
		}
		if v.GetNamespace() != "" {
			andFilters[secret.KeyNamespace] = v.GetNamespace()
		}
		if v.GetName() != "" {
			andFilters[secret.KeyName] = v.GetName()
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

func (s *Store) limit(charts ...*chart.Chart) int {
	limit := 0
	for _, v := range charts {
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
		out:    make(chan resource.Event),
	}

	go func() {
		defer close(s.out)

		for s.stream.Next(s.ctx) {
			var doc changeDocument
			if err := s.stream.Decode(&doc); err != nil {
				continue
			}

			var op resource.EventOP
			switch doc.OperationType {
			case "insert":
				op = resource.EventStore
			case "update":
				op = resource.EventSwap
			case "delete":
				op = resource.EventDelete
			default:
				continue
			}

			event := resource.Event{
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
func (s *Stream) Next() <-chan resource.Event {
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
