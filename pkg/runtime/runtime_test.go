package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestRuntime_LookupByID(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st := spec.NewStore()

	r := New(Config{
		Scheme: s,
		Store:  st,
	})

	defer r.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	_, _ = st.Store(ctx, meta)

	n, err := r.LookupByID(ctx, meta.GetID())
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.Equal(t, meta.GetID(), n.ID())
}

func TestRuntime_LookupByName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st := spec.NewStore()

	r := New(Config{
		Scheme: s,
		Store:  st,
	})

	defer r.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
		Name: faker.Word(),
	}

	_, _ = st.Store(ctx, meta)

	n, err := r.LookupByName(ctx, meta.GetName())
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.Equal(t, meta.GetID(), n.ID())
}

func TestRuntime_Insert(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st := spec.NewStore()

	r := New(Config{
		Scheme: s,
		Store:  st,
	})

	defer r.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	n1, err := r.Insert(ctx, meta)
	assert.NoError(t, err)
	assert.Equal(t, meta.GetID(), n1.ID())

	n2, err := r.LookupByID(ctx, meta.GetID())
	assert.NoError(t, err)
	assert.Equal(t, n1, n2)
}

func TestRuntime_Free(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st := spec.NewStore()

	r := New(Config{
		Scheme: s,
		Store:  st,
	})

	defer r.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	_, _ = r.Insert(ctx, meta)

	n1, err := r.LookupByID(ctx, meta.GetID())
	assert.NoError(t, err)
	assert.Equal(t, meta.GetID(), n1.ID())

	ok, err := r.Free(ctx, meta)
	assert.NoError(t, err)
	assert.True(t, ok)

	n2, err := r.LookupByID(ctx, meta.GetID())
	assert.NoError(t, err)
	assert.Nil(t, n2)
}

func TestRuntime_Load(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st := spec.NewStore()

	r := New(Config{
		Scheme: s,
		Store:  st,
	})
	defer r.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	_, _ = st.Store(ctx, meta)

	r.Load(ctx)

	n, err := r.LookupByID(ctx, meta.GetID())
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestRuntime_Listen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st := spec.NewStore()

	r := New(Config{
		Scheme: s,
		Store:  st,
	})

	defer r.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	_, _ = st.Store(ctx, meta)

	go r.Listen(ctx)

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
				if n, _ := r.LookupByID(ctx, meta.GetID()); n != nil {
					return
				}
			}
		}
	}()
}

func BenchmarkNewRuntime(b *testing.B) {
	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st := spec.NewStore()

	for i := 0; i < b.N; i++ {
		r := New(Config{
			Scheme: s,
			Store:  st,
		})
		r.Close()
	}
}
