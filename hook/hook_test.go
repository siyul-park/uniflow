package hook

import (
	"testing"

	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/spec"
	"github.com/siyul-park/uniflow/symbol"
	"github.com/stretchr/testify/assert"
)

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
