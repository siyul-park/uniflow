package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/value"
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

	specStore := spec.NewStore()
	valueStore := value.NewStore()

	r := New(Config{
		Scheme:     s,
		SpecStore:  specStore,
		ValueStore: valueStore,
	})
	defer r.Close()

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	_, _ = specStore.Store(ctx, meta)

	err := r.Load(ctx)
	assert.NoError(t, err)
}

func TestRuntime_Reconcile(t *testing.T) {
	t.Run("Spec", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		s.AddKnownType(kind, &spec.Meta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		specStore := spec.NewStore()
		valueStore := value.NewStore()

		h := hook.New()
		symbols := make(chan *symbol.Symbol)

		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			symbols <- sb
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadFunc(func(sb *symbol.Symbol) error {
			symbols <- sb
			return nil
		}))

		r := New(Config{
			Scheme:     s,
			Hook:       h,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})
		defer r.Close()

		err := r.Watch(ctx)
		assert.NoError(t, err)

		go r.Reconcile(ctx)

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
		}

		specStore.Store(ctx, meta)

		select {
		case sb := <-symbols:
			assert.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
		}

		specStore.Delete(ctx, meta)

		select {
		case sb := <-symbols:
			assert.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
		}
	})

	t.Run("Value", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		s.AddKnownType(kind, &spec.Meta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		specStore := spec.NewStore()
		valueStore := value.NewStore()

		h := hook.New()
		symbols := make(chan *symbol.Symbol)

		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			symbols <- sb
			return nil
		}))
		h.AddUnloadHook(symbol.UnloadFunc(func(sb *symbol.Symbol) error {
			symbols <- sb
			return nil
		}))

		r := New(Config{
			Scheme:     s,
			Hook:       h,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})

		err := r.Watch(ctx)
		assert.NoError(t, err)

		go r.Reconcile(ctx)

		scrt := &value.Value{
			ID:   uuid.Must(uuid.NewV7()),
			Data: faker.UUIDHyphenated(),
		}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Env: map[string]spec.Value{
				"key": {
					ID:   scrt.GetID(),
					Data: "{{ . }}",
				},
			},
		}

		specStore.Store(ctx, meta)
		valueStore.Store(ctx, scrt)

		select {
		case sb := <-symbols:
			assert.Equal(t, meta.GetID(), sb.ID())
			assert.Equal(t, scrt.Data, sb.Env()["key"].Data)
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
		}

		valueStore.Delete(ctx, scrt)

		select {
		case sb := <-symbols:
			assert.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
		}
	})
}

func BenchmarkRuntime_Reconcile(b *testing.B) {
	b.Run("Spec", func(b *testing.B) {
		ctx := context.TODO()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		s.AddKnownType(kind, &spec.Meta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		specStore := spec.NewStore()
		valueStore := value.NewStore()

		h := hook.New()
		symbols := make(chan *symbol.Symbol)

		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			symbols <- sb
			return nil
		}))

		r := New(Config{
			Scheme:     s,
			Hook:       h,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})
		defer r.Close()

		err := r.Watch(ctx)
		assert.NoError(b, err)

		go r.Reconcile(ctx)

		for i := 0; i < b.N; i++ {
			meta := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: resource.DefaultNamespace,
			}

			specStore.Store(ctx, meta)

			select {
			case <-symbols:
			case <-ctx.Done():
			}
		}
	})

	b.Run("Value", func(b *testing.B) {
		ctx := context.TODO()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		s.AddKnownType(kind, &spec.Meta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		specStore := spec.NewStore()
		valueStore := value.NewStore()

		h := hook.New()
		symbols := make(chan *symbol.Symbol)

		h.AddLoadHook(symbol.LoadFunc(func(sb *symbol.Symbol) error {
			symbols <- sb
			return nil
		}))

		r := New(Config{
			Scheme:     s,
			Hook:       h,
			SpecStore:  specStore,
			ValueStore: valueStore,
		})

		err := r.Watch(ctx)
		assert.NoError(b, err)

		go r.Reconcile(ctx)

		for i := 0; i < b.N; i++ {
			scrt := &value.Value{
				ID:   uuid.Must(uuid.NewV7()),
				Data: faker.UUIDHyphenated(),
			}
			meta := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: resource.DefaultNamespace,
				Env: map[string]spec.Value{
					"key": {
						ID:   scrt.GetID(),
						Data: "{{ . }}",
					},
				},
			}

			specStore.Store(ctx, meta)
			valueStore.Store(ctx, scrt)

			select {
			case <-symbols:
			case <-ctx.Done():
			}
		}
	})
}
