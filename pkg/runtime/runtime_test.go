package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestRuntime_Lookup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	db := memdb.New(faker.UUIDHyphenated())

	st, _ := scheme.NewStorage(ctx, scheme.StorageConfig{
		Scheme:   s,
		Database: db,
	})

	r, _ := New(ctx, Config{
		Scheme:   s,
		Database: db,
	})
	defer r.Close()

	spec := &scheme.SpecMeta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	_, _ = st.InsertOne(ctx, spec)

	n, err := r.Lookup(ctx, spec.GetID())
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.Equal(t, spec.GetID(), n.ID())
}

func TestRuntime_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	db := memdb.New(faker.UUIDHyphenated())

	st, _ := scheme.NewStorage(ctx, scheme.StorageConfig{
		Scheme:   s,
		Database: db,
	})

	r, _ := New(ctx, Config{
		Scheme:   s,
		Database: db,
	})
	defer r.Close()

	spec := &scheme.SpecMeta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	_, _ = st.InsertOne(ctx, spec)

	go r.Start(ctx)

	func() {
		ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()

		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				assert.NoError(t, ctx.Err())
				return
			case <-ticker.C:
				if n, _ := r.Lookup(ctx, spec.GetID()); n != nil {
					return
				}
			}
		}
	}()
}

func BenchmarkNewRuntime(b *testing.B) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	db := memdb.New(faker.UUIDHyphenated())

	for i := 0; i < b.N; i++ {
		r, _ := New(ctx, Config{
			Scheme:   s,
			Database: db,
		})
		r.Close()
	}
}
