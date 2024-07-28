package secret

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	_ "github.com/siyul-park/uniflow/driver/mongo/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/secret"
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
	out    chan secret.Event
}

type changeDocument struct {
	OperationType string `bson:"operationType"`
	DocumentKey   struct {
		ID uuid.UUID `bson:"_id"`
	} `bson:"documentKey"`
}

var _ secret.Store = (*Store)(nil)
var _ secret.Stream = (*Stream)(nil)

// NewStore creates a new Store with the specified MongoDB collection.
func NewStore(collection *mongo.Collection) *Store {
	return &Store{collection: collection}
}

// Index ensures the collection has the required indexes and updates them if necessary.
func (s *Store) Index(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: secret.KeyNamespace, Value: 1},
				{Key: secret.KeyName, Value: 1},
			},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{
				secret.KeyName: bson.M{"$exists": true},
			}),
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Watch returns a Stream that monitors changes matching the specified filter.
func (s *Store) Watch(ctx context.Context, secrets ...*secret.Secret) (secret.Stream, error) {
	filter := s.filter(secrets...)

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
func (s *Store) Load(ctx context.Context, secrets ...*secret.Secret) ([]*secret.Secret, error) {
	filter := s.filter(secrets...)
	limit := int64(s.limit(secrets...))

	cursor, err := s.collection.Find(ctx, filter, &options.FindOptions{
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []*secret.Secret
	for cursor.Next(ctx) {
		secret := &secret.Secret{}
		if err := cursor.Decode(secret); err != nil {
			return nil, err
		}
		result = append(result, secret)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// Store saves the given Specs into the database.
func (s *Store) Store(ctx context.Context, secrets ...*secret.Secret) (int, error) {
	var docs []any
	for _, v := range secrets {
		if v.GetID() == uuid.Nil {
			v.SetID(uuid.Must(uuid.NewV7()))
		}
		if v.GetNamespace() == "" {
			v.SetNamespace(secret.DefaultNamespace)
		}

		docs = append(docs, v)
	}

	res, err := s.collection.InsertMany(ctx, docs)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return 0, errors.WithMessage(secret.ErrDuplicatedKey, err.Error())
		}
		return 0, err
	}
	return len(res.InsertedIDs), nil
}

// Swap updates existing Specs in the database with the provided data.
func (s *Store) Swap(ctx context.Context, secrets ...*secret.Secret) (int, error) {
	ids := make([]uuid.UUID, len(secrets))
	for i, secret := range secrets {
		if secret.GetID() == uuid.Nil {
			secret.SetID(uuid.Must(uuid.NewV7()))
		}
		ids[i] = secret.GetID()
	}

	filter := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	exists := make(map[uuid.UUID]*secret.Secret)
	for cursor.Next(ctx) {
		secret := &secret.Secret{}
		if err := cursor.Decode(secret); err != nil {
			return 0, err
		}
		exists[secret.GetID()] = secret
	}

	count := 0
	for _, secret := range secrets {
		exist, ok := exists[secret.GetID()]
		if !ok {
			continue
		}

		secret.SetNamespace(exist.GetNamespace())

		filter := bson.M{"_id": secret.GetID()}
		update := bson.M{"$set": secret}

		res, err := s.collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return 0, err
		}
		count += int(res.MatchedCount)
	}
	return count, nil
}

// Delete removes Specs from the store based on the provided criteria.
func (s *Store) Delete(ctx context.Context, secrets ...*secret.Secret) (int, error) {
	filter := s.filter(secrets...)
	res, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int(res.DeletedCount), nil
}

func (s *Store) filter(secrets ...*secret.Secret) bson.M {
	var orFilters []bson.M
	for _, v := range secrets {
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

func (s *Store) limit(secrets ...*secret.Secret) int {
	limit := 0
	for _, v := range secrets {
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
		out:    make(chan secret.Event),
	}

	go func() {
		defer close(s.out)

		for s.stream.Next(s.ctx) {
			var doc changeDocument
			if err := s.stream.Decode(&doc); err != nil {
				continue
			}

			var op secret.EventOP
			switch doc.OperationType {
			case "insert":
				op = secret.EventStore
			case "update":
				op = secret.EventSwap
			case "delete":
				op = secret.EventDelete
			default:
				continue
			}

			event := secret.Event{
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
func (s *Stream) Next() <-chan secret.Event {
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