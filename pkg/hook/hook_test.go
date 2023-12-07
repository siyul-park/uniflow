package hook

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestHook_LoadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(nil)

	count := 0
	h := symbol.LoadHookFunc(func(_ node.Node) error {
		count += 1
		return nil
	})

	hooks.AddLoadHook(h)

	err := hooks.Load(n)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestHook_UnloadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(nil)

	count := 0
	h := symbol.UnloadHookFunc(func(_ node.Node) error {
		count += 1
		return nil
	})

	hooks.AddUnloadHook(h)

	err := hooks.Unload(n)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
