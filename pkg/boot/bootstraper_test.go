package boot

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/loader"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestBootstraper_Boot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	s := scheme.New()
	kind := faker.UUIDHyphenated()

	s.AddKnownType(kind, &spec.Meta{})
	s.AddCodec(kind, scheme.CodecFunc(func(spec spec.Spec) (node.Node, error) {
		return node.NewOneToOneNode(nil), nil
	}))

	st, _ := store.New(ctx, store.Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	tb := symbol.NewTable(s)
	defer tb.Clear()

	l := loader.New(loader.Config{
		Store: st,
		Table: tb,
	})

	r := loader.NewReconciler(loader.ReconcilerConfig{
		Namespace: spec.DefaultNamespace,
		Store:     st,
		Loader:    l,
	})
	defer r.Close()

	count := 0
	h := BootHookFunc(func(_ context.Context) error {
		count++
		return nil
	})

	b := NewBootstraper(BootstraperConfig{
		Loader:     l,
		Reconciler: r,
		BootHooks:  []BootHook{h},
	})

	err := b.Boot(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
