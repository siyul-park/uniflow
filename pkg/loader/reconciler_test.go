package loader

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestReconciler_Reconcile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := spec.NewScheme()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, spec.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := spec.NewStorage(ctx, spec.StorageConfig{
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
		Namespace: spec.DefaultNamespace,
		Storage:   st,
		Loader:    ld,
	})
	defer r.Close()

	err := r.Watch(ctx)
	assert.NoError(t, err)

	go r.Reconcile(ctx)

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
	}

	st.InsertOne(ctx, meta)

	func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				assert.NoError(t, ctx.Err())
				return
			default:
				if sym, ok := tb.LookupByID(meta.GetID()); ok {
					assert.Equal(t, meta.GetID(), sym.ID())
					return
				}
			}
		}
	}()
}
