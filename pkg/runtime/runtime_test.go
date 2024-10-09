package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
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

	chartStore := chart.NewStore()
	specStore := spec.NewStore()
	secretStore := secret.NewStore()

	r := New(Config{
		Scheme:      s,
		ChartStore:  chartStore,
		SpecStore:   specStore,
		SecretStore: secretStore,
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
	t.Run("Chart", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		chartStore := chart.NewStore()
		specStore := spec.NewStore()
		secretStore := secret.NewStore()

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
			Scheme:      s,
			Hook:        h,
			ChartStore:  chartStore,
			SpecStore:   specStore,
			SecretStore: secretStore,
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
		chrt := &chart.Chart{
			ID:        uuid.Must(uuid.NewV7()),
			Namespace: resource.DefaultNamespace,
			Name:      kind,
		}

		specStore.Store(ctx, meta)
		chartStore.Store(ctx, chrt)

		select {
		case sb := <-symbols:
			assert.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
			return
		}

		chartStore.Delete(ctx, chrt)

		select {
		case sb := <-symbols:
			assert.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
			return
		}
	})

	t.Run("Spec", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		s.AddKnownType(kind, &spec.Meta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		chartStore := chart.NewStore()
		specStore := spec.NewStore()
		secretStore := secret.NewStore()

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
			Scheme:      s,
			Hook:        h,
			ChartStore:  chartStore,
			SpecStore:   specStore,
			SecretStore: secretStore,
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
			return
		}

		specStore.Delete(ctx, meta)

		select {
		case sb := <-symbols:
			assert.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
			return
		}
	})

	t.Run("Secret", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		s.AddKnownType(kind, &spec.Meta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		chartStore := chart.NewStore()
		specStore := spec.NewStore()
		secretStore := secret.NewStore()

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
			Scheme:      s,
			Hook:        h,
			ChartStore:  chartStore,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})

		err := r.Watch(ctx)
		assert.NoError(t, err)

		go r.Reconcile(ctx)

		scrt := &secret.Secret{
			ID:   uuid.Must(uuid.NewV7()),
			Data: faker.Word(),
		}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Env: map[string][]spec.Value{
				"key": {
					{
						ID:    scrt.GetID(),
						Value: "{{ . }}",
					},
				},
			},
		}

		specStore.Store(ctx, meta)
		secretStore.Store(ctx, scrt)

		select {
		case sb := <-symbols:
			assert.Equal(t, meta.GetID(), sb.ID())
			assert.Equal(t, scrt.Data, sb.Env()["key"][0].Value)
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
			return
		}

		secretStore.Delete(ctx, scrt)

		select {
		case sb := <-symbols:
			assert.Equal(t, meta.GetID(), sb.ID())
		case <-ctx.Done():
			assert.NoError(t, ctx.Err())
			return
		}
	})
}
