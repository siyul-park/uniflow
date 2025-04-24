package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/hook"
	"github.com/siyul-park/uniflow/meta"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/spec"
	"github.com/siyul-park/uniflow/store"
	"github.com/siyul-park/uniflow/symbol"
	"github.com/siyul-park/uniflow/value"
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
	defer r.Close(ctx)

	meta := &spec.Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: kind,
	}

	err := specStore.Insert(ctx, []any{meta})
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
		defer r.Close(ctx)

		err := r.Watch(ctx)
		require.NoError(t, err)

		go r.Reconcile(ctx)

		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: meta.DefaultNamespace,
		}

		err = specStore.Insert(ctx, []any{meta})
		require.NoError(t, err)

		select {
		case sb := <-symbols:
			require.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			require.NoError(t, ctx.Err())
		}

		_, err = specStore.Delete(ctx, map[string]any{spec.KeyID: meta.ID})
		require.NoError(t, err)

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
		defer r.Close(ctx)

		err := r.Watch(ctx)
		require.NoError(t, err)

		go r.Reconcile(ctx)

		val := &value.Value{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: meta.DefaultNamespace,
			Data:      faker.UUIDHyphenated(),
		}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: meta.DefaultNamespace,
			Env: map[string]spec.Value{
				"key": {
					ID:   val.GetID(),
					Data: "{{ . }}",
				},
			},
		}

		err = specStore.Insert(ctx, []any{meta})
		require.NoError(t, err)

		err = valueStore.Insert(ctx, []any{val})
		require.NoError(t, err)

		select {
		case sb := <-symbols:
			require.Equal(t, meta.GetID(), sb.ID())
			require.Equal(t, val.Data, sb.Env()["key"].Data)
		case <-ctx.Done():
			require.NoError(t, ctx.Err())
		}

		_, err = valueStore.Delete(ctx, map[string]any{value.KeyID: val.ID})
		require.NoError(t, err)

		select {
		case sb := <-symbols:
			require.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			require.NoError(t, ctx.Err())
		}

		go func() {
			for range symbols {
			}
		}()
	})
}
