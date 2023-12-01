package storage

import (
	"context"
	"reflect"
	"sync"

	"github.com/oklog/ulid/v2"
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

// CollectionNodes is the name of the nodes collection in the storage.
const CollectionNodes = "nodes"

var (
	indexesNode = []database.IndexModel{
		{
			Name:    "namespace_name",
			Keys:    []string{scheme.KeyNamespace, scheme.KeyName},
			Unique:  true,
			Partial: database.Where(scheme.KeyName).NE(primitive.NewString("")).And(database.Where(scheme.KeyName).IsNotNull()),
		},
	}
)

// New creates a new Storage instance.
func New(ctx context.Context, config Config) (*Storage, error) {
	scheme := config.Scheme
	db := config.Database

	nodes, err := db.Collection(ctx, CollectionNodes)
	if err != nil {
		return nil, err
	}

	s := &Storage{
		scheme: scheme,
		nodes:  nodes,
	}

	if exists, err := s.nodes.Indexes().List(ctx); err != nil {
		return nil, err
	} else {
		for _, index := range indexesNode {
			index = database.IndexModel{
				Name:    index.Name,
				Keys:    index.Keys,
				Unique:  index.Unique,
				Partial: index.Partial,
			}

			var ok bool
			for _, i := range exists {
				if i.Name == index.Name {
					if reflect.DeepEqual(i, index) {
						s.nodes.Indexes().Drop(ctx, i.Name)
					}
					break
				}
			}
			if ok {
				continue
			}
			s.nodes.Indexes().Create(ctx, index)
		}
	}

	return s, nil
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
	return NewStream(stream), nil
}

// InsertOne inserts a single scheme.Spec and returns its ID.
func (s *Storage) InsertOne(ctx context.Context, spec scheme.Spec) (ulid.ULID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	unstructured := scheme.NewUnstructured(nil)

	if err := unstructured.Marshal(spec); err != nil {
		return ulid.ULID{}, err
	}
	if unstructured.GetNamespace() == "" {
		unstructured.SetNamespace(scheme.DefaultNamespace)
	}
	if unstructured.GetID() == (ulid.ULID{}) {
		unstructured.SetID(ulid.Make())
	}

	if err := s.validate(unstructured); err != nil {
		return ulid.ULID{}, err
	}

	var id ulid.ULID
	if pk, err := s.nodes.InsertOne(ctx, unstructured.Doc()); err != nil {
		return ulid.ULID{}, err
	} else if err := primitive.Unmarshal(pk, &id); err != nil {
		_, _ = s.nodes.DeleteOne(ctx, database.Where(scheme.KeyID).EQ(pk))
		return ulid.ULID{}, err
	} else {
		return id, nil
	}
}

// InsertMany inserts multiple scheme.Spec instances and returns their IDs.
func (s *Storage) InsertMany(ctx context.Context, objs []scheme.Spec) ([]ulid.ULID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var docs []*primitive.Map
	for _, spec := range objs {
		unstructured := scheme.NewUnstructured(nil)

		if err := unstructured.Marshal(spec); err != nil {
			return nil, err
		}
		if unstructured.GetNamespace() == "" {
			unstructured.SetNamespace(scheme.DefaultNamespace)
		}
		if unstructured.GetID() == (ulid.ULID{}) {
			unstructured.SetID(ulid.Make())
		}

		if err := s.validate(unstructured); err != nil {
			return nil, err
		}

		docs = append(docs, unstructured.Doc())
	}

	ids := make([]ulid.ULID, 0)
	if pks, err := s.nodes.InsertMany(ctx, docs); err != nil {
		return nil, err
	} else if err := primitive.Unmarshal(primitive.NewSlice(pks...), &ids); err != nil {
		_, _ = s.nodes.DeleteMany(ctx, database.Where(scheme.KeyID).IN(pks...))
		return nil, err
	} else {
		return ids, nil
	}
}

// UpdateOne updates a single scheme.Spec and returns success or failure.
func (s *Storage) UpdateOne(ctx context.Context, spec scheme.Spec) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	unstructured := scheme.NewUnstructured(nil)

	if err := unstructured.Marshal(spec); err != nil {
		return false, err
	}
	if unstructured.GetNamespace() == "" {
		unstructured.SetNamespace(scheme.DefaultNamespace)
	}
	if unstructured.GetID() == (ulid.ULID{}) {
		return false, nil
	}

	if err := s.validate(unstructured); err != nil {
		return false, err
	}

	filter, _ := Where[ulid.ULID](scheme.KeyID).EQ(unstructured.GetID()).Encode()
	return s.nodes.UpdateOne(ctx, filter, unstructured.Doc())
}

// UpdateMany updates multiple scheme.Spec instances and returns the number of successes.
func (s *Storage) UpdateMany(ctx context.Context, objs []scheme.Spec) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var unstructureds []*scheme.Unstructured
	for _, spec := range objs {
		unstructured := scheme.NewUnstructured(nil)

		if err := unstructured.Marshal(spec); err != nil {
			return 0, err
		}
		if unstructured.GetNamespace() == "" {
			unstructured.SetNamespace(scheme.DefaultNamespace)
		}
		if unstructured.GetID() == (ulid.ULID{}) {
			continue
		}

		if err := s.validate(unstructured); err != nil {
			return 0, err
		}

		unstructureds = append(unstructureds, unstructured)
	}

	count := 0
	for _, unstructured := range unstructureds {
		filter, _ := Where[ulid.ULID](scheme.KeyID).EQ(unstructured.GetID()).Encode()
		if ok, err := s.nodes.UpdateOne(ctx, filter, unstructured.Doc()); err != nil {
			return count, err
		} else if ok {
			count += 1
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

	if doc, err := s.nodes.FindOne(ctx, f, options...); err != nil {
		return nil, err
	} else if doc != nil {
		unstructured := scheme.NewUnstructured(doc)
		if spec, ok := s.scheme.New(unstructured.GetKind()); !ok {
			return unstructured, nil
		} else if err := unstructured.Unmarshal(spec); err != nil {
			return nil, err
		} else {
			return spec, nil
		}
	}

	return nil, nil
}

// FindMany returns multiple scheme.Spec instances matched by the filter.
func (s *Storage) FindMany(ctx context.Context, filter *Filter, options ...*database.FindOptions) ([]scheme.Spec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := filter.Encode()
	if err != nil {
		return nil, err
	}

	var specs []scheme.Spec
	if docs, err := s.nodes.FindMany(ctx, f, options...); err != nil {
		return nil, err
	} else {
		for _, doc := range docs {
			if doc == nil {
				continue
			}
			unstructured := scheme.NewUnstructured(doc)
			if spec, ok := s.scheme.New(unstructured.GetKind()); !ok {
				specs = append(specs, unstructured)
			} else if err := unstructured.Unmarshal(spec); err != nil {
				return nil, err
			} else {
				specs = append(specs, spec)
			}
		}

		return specs, nil
	}
}

func (s *Storage) validate(unstructured *scheme.Unstructured) error {
	if spec, ok := s.scheme.New(unstructured.GetKind()); ok {
		if err := unstructured.Unmarshal(spec); err != nil {
			return err
		} else if n, err := s.scheme.Decode(spec); err != nil {
			return err
		} else {
			_ = n.Close()
		}
	}
	return nil
}
