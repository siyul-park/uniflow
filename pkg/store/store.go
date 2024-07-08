package store

import (
	"context"
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// Store manages the storage and retrieval of spec.Spec objects.
type Store struct {
	nodes database.Collection
	mu    sync.RWMutex
}

var indexes = []database.IndexModel{
	{
		Name: "kind",
		Keys: []string{spec.KeyKind},
	},
	{
		Name:    "namespace_name",
		Keys:    []string{spec.KeyNamespace, spec.KeyName},
		Unique:  true,
		Partial: database.Where(spec.KeyName).NotEqual(types.NewString("")).And(database.Where(spec.KeyName).IsNotNull()),
	},
}

// New creates a new Store instance.
func New(nodes database.Collection) *Store {
	return &Store{nodes: nodes}
}

// Index ensures that all collection indexes are up-to-date.
func (s *Store) Index(ctx context.Context) error {
	origins, err := s.nodes.Indexes().List(ctx)
	if err != nil {
		return err
	}

	for _, index := range indexes {
		for _, origin := range origins {
			if origin.Name == index.Name {
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
	}
	return nil
}

// Watch returns a Stream that monitors changes based on the provided filter.
func (s *Store) Watch(ctx context.Context, filter *Filter) (*Stream, error) {
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

// InsertOne inserts a single spec.Spec and returns its UUID.
func (s *Store) InsertOne(ctx context.Context, spc spec.Spec) (uuid.UUID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if spc.GetNamespace() == "" {
		spc.SetNamespace(spec.DefaultNamespace)
	}
	if spc.GetID() == (uuid.UUID{}) {
		spc.SetID(uuid.Must(uuid.NewV7()))
	}

	val, err := types.BinaryEncoder.Encode(spc)
	if err != nil {
		return uuid.UUID{}, err
	}

	doc, ok := val.(types.Map)
	if !ok {
		return uuid.UUID{}, errors.WithStack(encoding.ErrInvalidArgument)
	}

	pk, err := s.nodes.InsertOne(ctx, doc)
	if err != nil {
		return uuid.UUID{}, err
	}

	var id uuid.UUID
	if err := types.Decoder.Decode(pk, &id); err != nil {
		_, _ = s.nodes.DeleteOne(ctx, database.Where(spec.KeyID).Equal(pk))
		return uuid.UUID{}, err
	}
	return id, nil
}

// InsertMany inserts multiple spec.Spec instances and returns their UUIDs.
func (s *Store) InsertMany(ctx context.Context, spcs []spec.Spec) ([]uuid.UUID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var docs []types.Map
	for _, spc := range spcs {
		if spc.GetNamespace() == "" {
			spc.SetNamespace(spec.DefaultNamespace)
		}
		if spc.GetID() == (uuid.UUID{}) {
			spc.SetID(uuid.Must(uuid.NewV7()))
		}

		val, err := types.BinaryEncoder.Encode(spc)
		if err != nil {
			return nil, err
		}

		doc, ok := val.(types.Map)
		if !ok {
			return nil, errors.WithStack(encoding.ErrInvalidArgument)
		}

		docs = append(docs, doc)
	}

	pks, err := s.nodes.InsertMany(ctx, docs)
	if err != nil {
		return nil, err
	}

	var ids []uuid.UUID
	if err := types.Decoder.Decode(types.NewSlice(pks...), &ids); err != nil {
		_, _ = s.nodes.DeleteMany(ctx, database.Where(spec.KeyID).In(pks...))
		return nil, err
	}
	return ids, nil
}

// UpdateOne updates a single spec.Spec and returns success or failure.
func (s *Store) UpdateOne(ctx context.Context, spc spec.Spec) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if spc.GetNamespace() == "" {
		spc.SetNamespace(spec.DefaultNamespace)
	}
	if spc.GetID() == (uuid.UUID{}) {
		spc.SetID(uuid.Must(uuid.NewV7()))
	}

	f, _ := Where[uuid.UUID](spec.KeyID).EQ(spc.GetID()).Encode()

	doc, err := s.nodes.FindOne(ctx, f)
	if err != nil {
		return false, err
	} else if doc == nil {
		return false, nil
	}

	unstructurd := &spec.Unstructured{}
	if err := types.Decoder.Decode(doc, unstructurd); err != nil {
		return false, err
	}

	if unstructurd.GetNamespace() != spc.GetNamespace() {
		return false, nil
	}

	val, err := types.BinaryEncoder.Encode(spc)
	if err != nil {
		return false, err
	}

	doc, ok := val.(types.Map)
	if !ok {
		return false, errors.WithStack(encoding.ErrInvalidArgument)
	}

	return s.nodes.UpdateOne(ctx, f, doc)
}

// UpdateMany updates multiple spec.Spec instances and returns the number of successes.
func (s *Store) UpdateMany(ctx context.Context, spcs []spec.Spec) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]uuid.UUID, len(spcs))
	for i, spc := range spcs {
		ids[i] = spc.GetID()
	}

	f, _ := Where[uuid.UUID](spec.KeyID).IN(ids...).Encode()

	docs, err := s.nodes.FindMany(ctx, f, &database.FindOptions{
		Limit: lo.ToPtr(len(ids)),
	})
	if err != nil {
		return 0, err
	}

	unstructurds := make(map[uuid.UUID]*spec.Unstructured, len(spcs))
	for _, doc := range docs {
		unstructurd := &spec.Unstructured{}
		if err := types.Decoder.Decode(doc, unstructurd); err != nil {
			return 0, err
		}
		unstructurds[unstructurd.ID] = unstructurd
	}

	count := 0
	for _, spc := range spcs {
		if spc.GetNamespace() == "" {
			spc.SetNamespace(spec.DefaultNamespace)
		}
		if spc.GetID() == (uuid.UUID{}) {
			spc.SetID(uuid.Must(uuid.NewV7()))
		}

		if unstructurd, ok := unstructurds[spc.GetID()]; !ok || unstructurd.GetNamespace() != spc.GetNamespace() {
			continue
		}

		val, err := types.BinaryEncoder.Encode(spc)
		if err != nil {
			return 0, err
		}

		doc, ok := val.(types.Map)
		if !ok {
			return 0, errors.WithStack(encoding.ErrInvalidArgument)
		}

		f, _ := Where[uuid.UUID](spec.KeyID).EQ(spc.GetID()).Encode()

		if ok, err := s.nodes.UpdateOne(ctx, f, doc); err != nil {
			return count, err
		} else if ok {
			count++
		}
	}
	return count, nil
}

// DeleteOne deletes a single spec.Spec and returns success or failure.
func (s *Store) DeleteOne(ctx context.Context, filter *Filter) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := filter.Encode()
	if err != nil {
		return false, err
	}

	return s.nodes.DeleteOne(ctx, f)
}

// DeleteMany deletes multiple spec.Spec instances and returns the number of successes.
func (s *Store) DeleteMany(ctx context.Context, filter *Filter) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := filter.Encode()
	if err != nil {
		return 0, err
	}

	return s.nodes.DeleteMany(ctx, f)
}

// FindOne returns a single spec.Spec matched by the filter.
func (s *Store) FindOne(ctx context.Context, filter *Filter, options ...*database.FindOptions) (spec.Spec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := filter.Encode()
	if err != nil {
		return nil, err
	}

	if doc, err := s.nodes.FindOne(ctx, f, options...); err != nil {
		return nil, err
	} else if doc == nil {
		return nil, nil
	} else {
		unstructurd := &spec.Unstructured{}
		if err := types.Decoder.Decode(doc, unstructurd); err != nil {
			return nil, err
		}
		return unstructurd, nil
	}
}

// FindMany returns multiple spec.Spec instances matched by the filter.
func (s *Store) FindMany(ctx context.Context, filter *Filter, options ...*database.FindOptions) ([]spec.Spec, error) {
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

	var spcs []spec.Spec
	for _, doc := range docs {
		if doc == nil {
			continue
		}
		unstructurd := &spec.Unstructured{}
		if err := types.Decoder.Decode(doc, unstructurd); err != nil {
			return nil, err
		}
		spcs = append(spcs, unstructurd)
	}
	return spcs, nil
}
