package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestReconciler_Reconcile(t *testing.T) {
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

		symbolTable := symbol.NewTable()
		defer symbolTable.Clear()

		symbolLoader := symbol.NewLoader(symbol.LoaderConfig{
			Table:       symbolTable,
			Scheme:      s,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})

		reconciler := NewReconiler(ReconcilerConfig{
			Scheme:       s,
			SpecStore:    specStore,
			SecretStore:  secretStore,
			SymbolTable:  symbolTable,
			SymbolLoader: symbolLoader,
		})
		defer reconciler.Close()

		err := reconciler.Watch(ctx)
		assert.NoError(t, err)

		go reconciler.Reconcile(ctx)

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
					if sb, ok := symbolTable.Lookup(meta.GetID()); ok {
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
					if _, ok := symbolTable.Lookup(meta.GetID()); !ok {
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

		symbolTable := symbol.NewTable()
		defer symbolTable.Clear()

		symbolLoader := symbol.NewLoader(symbol.LoaderConfig{
			Table:       symbolTable,
			Scheme:      s,
			SpecStore:   specStore,
			SecretStore: secretStore,
		})

		reconciler := NewReconiler(ReconcilerConfig{
			Scheme:       s,
			SpecStore:    specStore,
			SecretStore:  secretStore,
			SymbolTable:  symbolTable,
			SymbolLoader: symbolLoader,
		})
		defer reconciler.Close()

		err := reconciler.Watch(ctx)
		assert.NoError(t, err)

		go reconciler.Reconcile(ctx)

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
					if sb, ok := symbolTable.Lookup(meta.GetID()); ok {
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
					if _, ok := symbolTable.Lookup(meta.GetID()); !ok {
						return
					}
				}
			}
		}()
	})
}
