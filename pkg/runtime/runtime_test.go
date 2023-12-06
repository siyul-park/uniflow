package runtime

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestRuntime_Lookup(t *testing.T) {
	kind := faker.Word()

	sb := scheme.NewBuilder(func(s *scheme.Scheme) error {
		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
			return node.NewOneToOneNode(node.OneToOneNodeConfig{}), nil
		}))
		return nil
	})
	s, _ := sb.Build()

	db := memdb.New(faker.Word())

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: db,
	})

	r, _ := New(context.Background(), Config{
		Scheme:   s,
		Database: db,
	})
	defer func() { _ = r.Close(context.Background()) }()

	spec := &scheme.SpecMeta{
		ID:   ulid.Make(),
		Kind: kind,
	}

	_, _ = st.InsertOne(context.Background(), spec)

	n, err := r.Lookup(context.Background(), spec.ID)
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestRuntime_Free(t *testing.T) {
	kind := faker.Word()

	sb := scheme.NewBuilder(func(s *scheme.Scheme) error {
		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
			return node.NewOneToOneNode(node.OneToOneNodeConfig{}), nil
		}))
		return nil
	})
	s, _ := sb.Build()

	db := memdb.New(faker.Word())

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: db,
	})

	r, _ := New(context.Background(), Config{
		Scheme:   s,
		Database: db,
	})
	defer func() { _ = r.Close(context.Background()) }()

	spec := &scheme.SpecMeta{
		ID:   ulid.Make(),
		Kind: kind,
	}

	_, _ = st.InsertOne(context.Background(), spec)
	_, _ = r.Lookup(context.Background(), spec.ID)

	ok, err := r.Free(context.Background(), spec.ID)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestRuntime_Start(t *testing.T) {
	kind := faker.Word()

	sb := scheme.NewBuilder(func(s *scheme.Scheme) error {
		s.AddKnownType(kind, &scheme.SpecMeta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
			return node.NewOneToOneNode(node.OneToOneNodeConfig{}), nil
		}))
		return nil
	})
	s, _ := sb.Build()

	db := memdb.New(faker.Word())

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: db,
	})

	r, _ := New(context.Background(), Config{
		Scheme:   s,
		Database: db,
	})
	defer func() { _ = r.Close(context.Background()) }()

	spec := &scheme.SpecMeta{
		ID:   ulid.Make(),
		Kind: kind,
	}

	_, _ = st.InsertOne(context.Background(), spec)

	go func() {
		err := r.Start(context.Background())
		assert.NoError(t, err)
	}()
}
