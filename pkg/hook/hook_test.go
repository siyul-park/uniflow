package hook

import (
	"context"
	"testing"

	"github.com/siyul-park/uniflow/pkg/boot"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestHook_BootHook(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	hooks := New()

	count := 0
	h := boot.BootHookFunc(func(_ context.Context) error {
		count += 1
		return nil
	})

	hooks.AddBootHook(h)

	err := hooks.Boot(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestHook_LoadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(nil)

	count := 0
	h := symbol.LoadHookFunc(func(_ *symbol.Symbol) error {
		count += 1
		return nil
	})

	hooks.AddLoadHook(h)

	err := hooks.Load(symbol.New(&spec.Meta{}, n))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestHook_UnloadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(nil)

	count := 0
	h := symbol.UnloadHookFunc(func(_ *symbol.Symbol) error {
		count += 1
		return nil
	})

	hooks.AddUnloadHook(h)

	err := hooks.Unload(symbol.New(&spec.Meta{}, n))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
