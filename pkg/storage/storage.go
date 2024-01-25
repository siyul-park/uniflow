package storage

import (
	"context"
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// Config is a configuration struct for Storage.
type Config struct {
	Scheme   *scheme.Scheme
	Database database.Database
}

// Storage is responsible for storing scheme.Spec.
type Storage struct {
	scheme *scheme.Scheme
	nodes  database.Collection
	mu     sync.RWMutex
}

var indexes = []database.IndexModel{
	{
		Name:    "namespace_name",
		Keys:    []string{scheme.KeyNamespace, scheme.KeyName},
		Unique:  true,
		Partial: database.Where(scheme.KeyName).NE(primitive.NewString("")).And(database.Where(scheme.KeyName).IsNotNull()),
	},
}

// New creates a new Storage instance.
func New(ctx context.Context, config Config) (*Storage, error) {
	scheme := config.Scheme
	db := config.Database

	nodes, err := db.Collection(ctx, "nodes")
	if err != nil {
		return nil, err
	}

	if err := ensureIndexes(ctx, nodes, indexes); err != nil {
		return nil, err
	}

	return &Storage{
		scheme: scheme,
		nodes:  nodes,
	}, nil
}

// Watch returns a Stream to track changes based on the provided filter.
func (s *Storage) Watch(ctx context.Context, filter *Filter) (*Stream, error) {
	f, err := filter.Encode()
	if err != nil {
		return nil, err
	}

	stream, err := s.nodes.Watch(ctx, f)
	if err != nil {
		return nil, err
	}

	return newStream(stream), nil
}

// InsertOne inserts a single scheme.Spec and returns its ID.
func (s *Storage) InsertOne(ctx context.Context, spec scheme.Spec) (uuid.UUID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, err := s.specToDoc(spec)
	if err != nil {
		return uuid.UUID{}, err
	}

	pk, err := s.nodes.InsertOne(ctx, doc)
	if err != nil {
		return uuid.UUID{}, err
	}

	var id uuid.UUID
	if err := primitive.Unmarshal(pk, &id); err != nil {
		_, _ = s.nodes.DeleteOne(ctx, database.Where(scheme.KeyID).EQ(pk))
		return uuid.UUID{}, err
	}

	return id, nil
}

// InsertMany inserts multiple scheme.Spec instances and returns their IDs.
func (s *Storage) InsertMany(ctx context.Context, specs []scheme.Spec) ([]uuid.UUID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var docs []*primitive.Map
	for _, spec := range specs {
		doc, err := s.specToDoc(spec)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}

	pks, err := s.nodes.InsertMany(ctx, docs)
	if err != nil {
		return nil, err
	}

	var ids []uuid.UUID
	if err := primitive.Unmarshal(primitive.NewSlice(pks...), &ids); err != nil {
		_, _ = s.nodes.DeleteMany(ctx, database.Where(scheme.KeyID).IN(pks...))
		return nil, err
	}

	return ids, nil
}

// UpdateOne updates a single scheme.Spec and returns success or failure.
func (s *Storage) UpdateOne(ctx context.Context, spec scheme.Spec) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	unstructured, err := s.specToUnstructured(spec)
	if err != nil {
		return false, err
	}

	filter, _ := Where[uuid.UUID](scheme.KeyID).EQ(unstructured.GetID()).Encode()
	return s.nodes.UpdateOne(ctx, filter, unstructured.Doc())
}

// UpdateMany updates multiple scheme.Spec instances and returns the number of successes.
func (s *Storage) UpdateMany(ctx context.Context, specs []scheme.Spec) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var unstructureds []*scheme.Unstructured
	for _, spec := range specs {
		unstructured, err := s.specToUnstructured(spec)
		if err != nil {
			return 0, err
		}
		unstructureds = append(unstructureds, unstructured)
	}

	count := 0
	for _, unstructured := range unstructureds {
		filter, _ := Where[uuid.UUID](scheme.KeyID).EQ(unstructured.GetID()).Encode()
		if ok, err := s.nodes.UpdateOne(ctx, filter, unstructured.Doc()); err != nil {
			return count, err
		} else if ok {
			count++
		}
	}

	return count, nil
}

// DeleteOne deletes a single scheme.Spec and returns success or failure.
func (s *Storage) DeleteOne(ctx context.Context, filter *Filter) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := filter.Encode()
	if err != nil {
		return false, err
	}

	return s.nodes.DeleteOne(ctx, f)
}

// DeleteMany deletes multiple scheme.Spec instances and returns the number of successes.
func (s *Storage) DeleteMany(ctx context.Context, filter *Filter) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := filter.Encode()
	if err != nil {
		return 0, err
	}

	return s.nodes.DeleteMany(ctx, f)
}

// FindOne returns a single scheme.Spec matched by the filter.
func (s *Storage) FindOne(ctx context.Context, filter *Filter, options ...*database.FindOptions) (scheme.Spec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := filter.Encode()
	if err != nil {
		return nil, err
	}

	doc, err := s.nodes.FindOne(ctx, f, options...)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	return s.scheme.NewSpecWithDoc(doc)
}

// FindMany returns multiple scheme.Spec instances matched by the filter.
func (s *Storage) FindMany(ctx context.Context, filter *Filter, options ...*database.FindOptions) ([]scheme.Spec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := filter.Encode()
	if err != nil {
		return nil, err
	}

	docs, err := s.nodes.FindMany(ctx, f, options...)
	if err != nil {
		return nil, err
	}

	var specs []scheme.Spec
	for _, doc := range docs {
		if doc == nil {
			continue
		}

		if spec, err := s.scheme.NewSpecWithDoc(doc); err != nil {
			return nil, err
		} else {
			specs = append(specs, spec)
		}
	}

	return specs, nil
}

func (s *Storage) specToDoc(spec scheme.Spec) (*primitive.Map, error) {
	unstructured, err := s.specToUnstructured(spec)
	if err != nil {
		return nil, err
	}
	if unstructured.GetID() == (uuid.UUID{}) {
		unstructured.SetID(uuid.Must(uuid.NewV7()))
	}
	return unstructured.Doc(), nil
}

func (s *Storage) specToUnstructured(spec scheme.Spec) (*scheme.Unstructured, error) {
	unstructured := scheme.NewUnstructured(nil)
	if err := unstructured.Marshal(spec); err != nil {
		return nil, err
	}
	if unstructured.GetNamespace() == "" {
		unstructured.SetNamespace(scheme.DefaultNamespace)
	}
	if err := s.validate(unstructured); err != nil {
		return nil, err
	}
	return unstructured, nil
}

func (s *Storage) validate(spec scheme.Spec) error {
	n, err := s.scheme.Decode(spec)
	if err != nil {
		return err
	}
	return n.Close()
}

func ensureIndexes(ctx context.Context, coll database.Collection, indexes []database.IndexModel) error {
	origins, err := coll.Indexes().List(ctx)
	if err != nil {
		return err
	}

	for _, index := range indexes {
		var exists bool
		for _, origin := range origins {
			if origin.Name == index.Name {
				if !reflect.DeepEqual(origin, index) {
					coll.Indexes().Drop(ctx, origin.Name)
				} else {
					exists = true
				}
				break
			}
		}
		if !exists {
			coll.Indexes().Create(ctx, index)
		}
	}

	return nil
}
