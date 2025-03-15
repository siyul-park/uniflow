package runtime

import (
	"context"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/types"
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
	"github.com/stretchr/testify/require"
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

	specStore := store.New()
	valueStore := store.New()

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

	d, err := types.Marshal(meta)
	require.NoError(t, err)

	doc, ok := d.(types.Map)
	require.True(t, ok)

	err = specStore.Insert(ctx, []types.Map{doc})
	require.NoError(t, err)

	err = r.Load(ctx, nil)
	require.NoError(t, err)
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

		specStore := store.New()
		valueStore := store.New()

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
		require.NoError(t, err)

		go r.Reconcile(ctx)

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
		}

		d, err := types.Marshal(meta)
		require.NoError(t, err)

		doc, ok := d.(types.Map)
		require.True(t, ok)

		err = specStore.Insert(ctx, []types.Map{doc})
		require.NoError(t, err)

		select {
		case sb := <-symbols:
			require.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			require.NoError(t, ctx.Err())
		}

		specStore.Delete(ctx, store.Where(spec.KeyID).Equal(doc.Get(types.NewString(spec.KeyID))))

		select {
		case sb := <-symbols:
			require.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			require.NoError(t, ctx.Err())
		}
	})

	t.Run("Value", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancel()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		s.AddKnownType(kind, &spec.Meta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		specStore := store.New()
		valueStore := store.New()

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
		require.NoError(t, err)

		go r.Reconcile(ctx)

		val := &value.Value{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: resource.DefaultNamespace,
			Data:      faker.UUIDHyphenated(),
		}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Env: map[string]spec.Value{
				"key": {
					ID:   val.GetID(),
					Data: "{{ . }}",
				},
			},
		}

		d, err := types.Marshal(meta)
		require.NoError(t, err)

		doc, ok := d.(types.Map)
		require.True(t, ok)

		err = specStore.Insert(ctx, []types.Map{doc})
		require.NoError(t, err)

		d, err = types.Marshal(val)
		require.NoError(t, err)

		doc, ok = d.(types.Map)
		require.True(t, ok)

		err = valueStore.Insert(ctx, []types.Map{doc})
		require.NoError(t, err)

		select {
		case sb := <-symbols:
			require.Equal(t, meta.GetID(), sb.ID())
			require.Equal(t, val.Data, sb.Env()["key"].Data)
		case <-ctx.Done():
			require.NoError(t, ctx.Err())
		}

		valueStore.Delete(ctx, store.Where(value.KeyID).Equal(doc.Get(types.NewString(value.KeyID))))

		select {
		case sb := <-symbols:
			require.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			require.NoError(t, ctx.Err())
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

		specStore := store.New()
		valueStore := store.New()

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
		require.NoError(b, err)

		go r.Reconcile(ctx)

		for i := 0; i < b.N; i++ {
			meta := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: resource.DefaultNamespace,
			}

			d, err := types.Marshal(meta)
			require.NoError(b, err)

			doc, ok := d.(types.Map)
			require.True(b, ok)

			err = specStore.Insert(ctx, []types.Map{doc})
			require.NoError(b, err)

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

		specStore := store.New()
		valueStore := store.New()

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
		require.NoError(b, err)

		go r.Reconcile(ctx)

		for i := 0; i < b.N; i++ {
			val := &value.Value{
				ID:   uuid.Must(uuid.NewV7()),
				Data: faker.UUIDHyphenated(),
			}
			meta := &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      kind,
				Namespace: resource.DefaultNamespace,
				Env: map[string]spec.Value{
					"key": {
						ID:   val.GetID(),
						Data: "{{ . }}",
					},
				},
			}

			d, err := types.Marshal(meta)
			require.NoError(b, err)

			doc, ok := d.(types.Map)
			require.True(b, ok)

			err = specStore.Insert(ctx, []types.Map{doc})
			require.NoError(b, err)

			d, err = types.Marshal(val)
			require.NoError(b, err)

			doc, ok = d.(types.Map)
			require.True(b, ok)

			err = valueStore.Insert(ctx, []types.Map{doc})
			require.NoError(b, err)

			select {
			case <-symbols:
			case <-ctx.Done():
			}
		}
	})
}
