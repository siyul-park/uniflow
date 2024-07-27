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
		Scheme:    s,
		SpecStore: st,
	})

	defer r.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	_, _ = st.Store(ctx, meta)

	symbols, err := r.Load(ctx, meta)
	assert.NoError(t, err)
	assert.Len(t, symbols, 1)
}

func TestRuntime_Store(t *testing.T) {
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
		Scheme:    s,
		SpecStore: st,
	})

	defer r.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	symbols, err := r.Store(ctx, meta)
	assert.NoError(t, err)
	assert.Len(t, symbols, 1)

	symbols, err = r.Load(ctx, meta)
	assert.NoError(t, err)
	assert.Len(t, symbols, 1)
}

func TestRuntime_Delete(t *testing.T) {
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
		Scheme:    s,
		SpecStore: st,
	})

	defer r.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	_, _ = r.Store(ctx, meta)

	symbols, err := r.Load(ctx, meta)
	assert.NoError(t, err)
	assert.Len(t, symbols, 1)

	count, err := r.Delete(ctx, meta)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	symbols, err = r.Load(ctx, meta)
	assert.NoError(t, err)
	assert.Len(t, symbols, 0)
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
		Scheme:    s,
		SpecStore: st,
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
				if symbols, _ := r.Load(ctx, meta); len(symbols) > 0 {
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
			Scheme:    s,
			SpecStore: st,
		})
		r.Close()
	}
}
