package store

import (
	"context"
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// Store is responsible for storing spec.Spec.
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
func New(ctx context.Context, nodes database.Collection) (*Store, error) {
	origins, err := nodes.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}

	for _, index := range indexes {
		var exists bool
		for _, origin := range origins {
			if origin.Name == index.Name {
				if !reflect.DeepEqual(origin, index) {
					nodes.Indexes().Drop(ctx, origin.Name)
				} else {
					exists = true
				}
				break
			}
		}
		if !exists {
			nodes.Indexes().Create(ctx, index)
		}
	}

	return &Store{nodes: nodes}, nil
}

// Watch returns a Stream to track changes based on the provided filter.
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

// InsertOne inserts a single spec.Spec and returns its ID.
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

// InsertMany inserts multiple spec.Spec instances and returns their IDs.
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

	val, err := types.BinaryEncoder.Encode(spc)
	if err != nil {
		return false, err
	}

	doc, ok := val.(types.Map)
	if !ok {
		return false, errors.WithStack(encoding.ErrInvalidArgument)
	}

	filter, _ := Where[uuid.UUID](spec.KeyID).EQ(spc.GetID()).Encode()

	return s.nodes.UpdateOne(ctx, filter, doc)
}

// UpdateMany updates multiple spec.Spec instances and returns the number of successes.
func (s *Store) UpdateMany(ctx context.Context, spcs []spec.Spec) (int, error) {
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
			return 0, err
		}

		doc, ok := val.(types.Map)
		if !ok {
			return 0, errors.WithStack(encoding.ErrInvalidArgument)
		}

		docs = append(docs, doc)
	}

	count := 0
	for i, doc := range docs {
		filter, _ := Where[uuid.UUID](spec.KeyID).EQ(spcs[i].GetID()).Encode()
		if ok, err := s.nodes.UpdateOne(ctx, filter, doc); err != nil {
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
		return spec.NewUnstructured(doc), nil
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
		spcs = append(spcs, spec.NewUnstructured(doc))
	}
	return spcs, nil
}
