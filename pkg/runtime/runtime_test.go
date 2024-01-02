package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestRuntime_Lookup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.Word()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	db := memdb.New(faker.Word())

	st, _ := storage.New(ctx, storage.Config{
		Scheme:   s,
		Database: db,
	})

	r, _ := New(ctx, Config{
		Scheme:   s,
		Database: db,
	})
	defer r.Close()

	spec := &scheme.SpecMeta{
		ID:   ulid.Make(),
		Kind: kind,
	}

	_, _ = st.InsertOne(ctx, spec)

	n, err := r.Lookup(ctx, spec.GetID())
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.Equal(t, spec.GetID(), n.ID())
}

func TestRuntime_Free(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.Word()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	db := memdb.New(faker.Word())

	st, _ := storage.New(ctx, storage.Config{
		Scheme:   s,
		Database: db,
	})

	r, _ := New(ctx, Config{
		Scheme:   s,
		Database: db,
	})
	defer r.Close()

	spec := &scheme.SpecMeta{
		ID:   ulid.Make(),
		Kind: kind,
	}

	_, _ = st.InsertOne(ctx, spec)
	_, _ = r.Lookup(ctx, spec.GetID())

	ok, err := r.Free(ctx, spec.GetID())
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestRuntime_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.Word()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	db := memdb.New(faker.Word())

	st, _ := storage.New(ctx, storage.Config{
		Scheme:   s,
		Database: db,
	})

	r, _ := New(ctx, Config{
		Scheme:   s,
		Database: db,
	})
	defer r.Close()

	spec := &scheme.SpecMeta{
		ID:   ulid.Make(),
		Kind: kind,
	}

	_, _ = st.InsertOne(ctx, spec)

	go func() {
		err := r.Start(ctx)
		assert.ErrorIs(t, context.Canceled, err)
	}()

	deadline := 200 * time.Millisecond
	tick := 5 * time.Millisecond

	limit := int(deadline.Milliseconds() / tick.Milliseconds())

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for i := 0; i < limit; i++ {
		<-ticker.C

		n, err := r.Lookup(ctx, spec.GetID())
		assert.NoError(t, err)
		if n != nil {
			return
		}
	}
}
