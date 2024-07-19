package spec

import (
	"context"
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/types"
)

// Store manages storage and retrieval of Spec objects in a database.
type Store struct {
	nodes database.Collection
	mu    sync.RWMutex
}

var indexes = []database.IndexModel{
	{
		Name: "kind",
		Keys: []string{KeyKind},
	},
	{
		Name:    "namespace_name",
		Keys:    []string{KeyNamespace, KeyName},
		Unique:  true,
		Partial: database.Where(KeyName).NotEqual(types.NewString("")).And(database.Where(KeyName).IsNotNull()),
	},
}

// NewStore creates a new Store with the specified database collection.
func NewStore(nodes database.Collection) *Store {
	return &Store{nodes: nodes}
}

// Index ensures the collection has the required indexes and updates them if necessary.
func (s *Store) Index(ctx context.Context) error {
	origins, err := s.nodes.Indexes().List(ctx)
	if err != nil {
		return err
	}

	for _, index := range indexes {
		found := false
		for _, origin := range origins {
			if origin.Name == index.Name {
				found = true
				if !reflect.DeepEqual(origin, index) {
					if err := s.nodes.Indexes().Drop(ctx, origin.Name); err != nil {
						return err
					}
					if err := s.nodes.Indexes().Create(ctx, index); err != nil {
						return err
					}
				}
				break
			}
		}
		if !found {
			if err := s.nodes.Indexes().Create(ctx, index); err != nil {
				return err
			}
		}
	}
	return nil
}

// Watch returns a Stream that monitors changes matching the specified filter.
func (s *Store) Watch(ctx context.Context, spec Spec) (*Stream, error) {
	filter := s.filter(spec)

	stream, err := s.nodes.Watch(ctx, filter)
	if err != nil {
		return nil, err
	}
	return newStream(stream), nil
}

// Load retrieves Specs from the store that match the given criteria.
func (s *Store) Load(ctx context.Context, specs ...Spec) ([]Spec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filter := s.filter(specs...)

	docs, err := s.nodes.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := make([]Spec, 0, len(docs))
	for _, doc := range docs {
		spec := &Unstructured{}
		if err := types.Decoder.Decode(doc, spec); err != nil {
			return nil, err
		}
		result = append(result, spec)
	}

	return result, nil
}

// Store saves the given Specs into the database.
func (s *Store) Store(ctx context.Context, specs ...Spec) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var docs []types.Map
	for _, spec := range specs {
		if spec.GetNamespace() == "" {
			spec.SetNamespace(DefaultNamespace)
		}
		if spec.GetID() == (uuid.UUID{}) {
			spec.SetID(uuid.Must(uuid.NewV7()))
		}

		val, err := types.BinaryEncoder.Encode(spec)
		if err != nil {
			return 0, err
		}

		doc, ok := val.(types.Map)
		if !ok {
			return 0, errors.WithStack(encoding.ErrUnsupportedValue)
		}

		docs = append(docs, doc)
	}

	pks, err := s.nodes.InsertMany(ctx, docs)
	return len(pks), err
}

// Swap updates existing Specs in the database with the provided data.
func (s *Store) Swap(ctx context.Context, specs ...Spec) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _, spec := range specs {
		if spec.GetNamespace() == "" {
			spec.SetNamespace(DefaultNamespace)
		}
		if spec.GetID() == (uuid.UUID{}) {
			spec.SetID(uuid.Must(uuid.NewV7()))
		}

		val, err := types.BinaryEncoder.Encode(spec)
		if err != nil {
			return 0, err
		}

		doc, ok := val.(types.Map)
		if !ok {
			return 0, errors.WithStack(encoding.ErrUnsupportedValue)
		}

		filter := database.Where(KeyID).Equal(types.NewBinary(spec.GetID().Bytes()))

		ok, err = s.nodes.UpdateOne(ctx, filter, doc)
		if err != nil {
			return 0, err
		}
		if ok {
			count++
		}
	}

	return count, nil
}

// Delete removes Specs from the store based on the provided criteria.
func (s *Store) Delete(ctx context.Context, specs ...Spec) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filter := s.filter(specs...)
	return s.nodes.DeleteMany(ctx, filter)
}

func (s *Store) filter(specs ...Spec) *database.Filter {
	var or []*database.Filter
	for _, spec := range specs {
		var and []*database.Filter
		if spec != nil {
			if spec.GetNamespace() != "" {
				and = append(and, database.Where(KeyNamespace).Equal(types.NewString(spec.GetNamespace())))
			}
			if spec.GetID() != (uuid.UUID{}) {
				and = append(and, database.Where(KeyID).Equal(types.NewBinary(spec.GetID().Bytes())))
			}
			if spec.GetName() != "" {
				and = append(and, database.Where(KeyName).Equal(types.NewString(spec.GetName())))
			}
		}
		if len(and) > 0 {
			if len(and) == 1 {
				or = append(or, and[0])
			} else {
				or = append(or, &database.Filter{OP: database.AND, Children: and})
			}
		}
	}

	if len(or) == 0 {
		return nil
	}
	if len(or) == 1 {
		return or[0]
	}
	return &database.Filter{OP: database.OR, Children: or}
}
