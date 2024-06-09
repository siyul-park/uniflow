package loader

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestReconciler_Reconcile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := scheme.NewStorage(ctx, scheme.StorageConfig{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	tb := symbol.NewTable(s)
	defer tb.Clear()

	ld := New(Config{
		Storage: st,
		Table:   tb,
	})

	r := NewReconciler(ReconcilerConfig{
		Namespace: scheme.DefaultNamespace,
		Storage:   st,
		Loader:    ld,
	})
	defer r.Close()

	err := r.Watch(ctx)
	assert.NoError(t, err)

	go r.Reconcile(ctx)

	spec := &scheme.SpecMeta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: scheme.DefaultNamespace,
	}

	st.InsertOne(ctx, spec)

	func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				assert.NoError(t, ctx.Err())
				return
			default:
				if sym, ok := tb.LookupByID(spec.GetID()); ok {
					assert.Equal(t, spec.GetID(), sym.ID())
					return
				}
			}
		}
	}()
}
