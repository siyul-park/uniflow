package loader

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestReconciler_Reconcile(t *testing.T) {
	s := scheme.New()

	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.Word()),
	})

	tb := symbol.NewTable()
	defer func() { _ = tb.Close() }()

	ld, _ := New(context.Background(), Config{
		Scheme:  s,
		Storage: st,
		Table:   tb,
	})

	r := NewReconciler(ReconcilerConfig{
		Remote: st,
		Loader: ld,
	})
	defer func() { _ = r.Close() }()

	err := r.Watch(context.Background())
	assert.NoError(t, err)

	go func() {
		err := r.Reconcile(context.Background())
		assert.NoError(t, err)
	}()

	kind := faker.Word()

	m := &scheme.SpecMeta{
		ID:        ulid.Make(),
		Kind:      kind,
		Namespace: scheme.NamespaceDefault,
	}

	codec := scheme.CodecFunc(func(spec scheme.Spec) (node.Node, error) {
		return node.NewOneToOneNode(node.OneToOneNodeConfig{ID: spec.GetID()}), nil
	})

	s.AddKnownType(kind, &scheme.SpecMeta{})
	s.AddCodec(kind, codec)

	st.InsertOne(context.Background(), m)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			assert.Fail(t, "timeout")
			return
		default:
			if _, ok := tb.Lookup(m.GetID()); ok {
				return
			}
		}
	}
}
