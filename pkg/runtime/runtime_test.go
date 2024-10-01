package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/agent"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
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

	specStore := spec.NewStore()
	secretStore := secret.NewStore()

	r := New(Config{
		Scheme:      s,
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
	t.Run("ReconcileLoadedSpec", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		s.AddKnownType(kind, &spec.Meta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		specStore := spec.NewStore()
		secretStore := secret.NewStore()

		h := hook.New()

		a := agent.New()
		defer a.Close()

		h.AddLoadHook(a)
		h.AddUnloadHook(a)

		r := New(Config{
			Scheme:      s,
			Hook:        h,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})
		defer r.Close()

		err := r.Watch(ctx)
		assert.NoError(t, err)

		go r.Reconcile(ctx)

		sec := &secret.Secret{
			ID:   uuid.Must(uuid.NewV7()),
			Data: faker.Word(),
		}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Env: map[string][]spec.Secret{
				"key": {
					{
						ID:    sec.GetID(),
						Value: "{{ . }}",
					},
				},
			},
		}

		secretStore.Store(ctx, sec)
		specStore.Store(ctx, meta)

		func() {
			for {
				select {
				case <-ctx.Done():
					assert.NoError(t, ctx.Err())
					return
				default:
					if sb := a.Symbol(meta.GetID()); sb != nil {
						assert.Equal(t, meta.GetID(), sb.ID())
						assert.Equal(t, sec.Data, sb.Env()["key"][0].Value)
						return
					}

				}
			}
		}()

		specStore.Delete(ctx, meta)

		func() {
			for {
				select {
				case <-ctx.Done():
					assert.NoError(t, ctx.Err())
					return
				default:
					if sb := a.Symbol(meta.GetID()); sb == nil {
						return
					}
				}
			}
		}()
	})

	t.Run("ReconcileLoadedSecret", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		s := scheme.New()
		kind := faker.UUIDHyphenated()

		s.AddKnownType(kind, &spec.Meta{})
		s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
			return node.NewOneToOneNode(nil), nil
		}))

		specStore := spec.NewStore()
		secretStore := secret.NewStore()

		h := hook.New()

		a := agent.New()
		defer a.Close()

		h.AddLoadHook(a)
		h.AddUnloadHook(a)

		r := New(Config{
			Scheme:      s,
			Hook:        h,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})

		err := r.Watch(ctx)
		assert.NoError(t, err)

		go r.Reconcile(ctx)

		sec := &secret.Secret{
			ID:   uuid.Must(uuid.NewV7()),
			Data: faker.Word(),
		}
		meta := &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      kind,
			Namespace: resource.DefaultNamespace,
			Env: map[string][]spec.Secret{
				"key": {
					{
						ID:    sec.GetID(),
						Value: "{{ . }}",
					},
				},
			},
		}

		specStore.Store(ctx, meta)
		secretStore.Store(ctx, sec)

		func() {
			for {
				select {
				case <-ctx.Done():
					assert.NoError(t, ctx.Err())
					return
				default:
					if sb := a.Symbol(meta.GetID()); sb != nil {
						if sec.Data == sb.Env()["key"][0].Value {
							return
						}
					}

				}
			}
		}()

		secretStore.Delete(ctx, sec)

		func() {
			for {
				select {
				case <-ctx.Done():
					assert.NoError(t, ctx.Err())
					return
				default:
					if sb := a.Symbol(meta.GetID()); sb == nil {
						return
					}
				}
			}
		}()
	})
}

func BenchmarkNewRuntime(b *testing.B) {
	kind := faker.UUIDHyphenated()

	s := scheme.New()
	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	specStore := spec.NewStore()
	secretStore := secret.NewStore()

	for i := 0; i < b.N; i++ {
		r := New(Config{
			Scheme:      s,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})
		r.Close()
	}
}
