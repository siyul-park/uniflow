package hook

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestHook_LoadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(node.OneToOneNodeConfig{})

	count := 0
	h := symbol.LoadHookFunc(func(_ node.Node) {
		count += 1
	})

	hooks.AddLoadHook(h)

	hooks.Load(n)
	assert.Equal(t, 1, count)
}

func TestHook_UnloadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(node.OneToOneNodeConfig{})

	count := 0
	h := symbol.UnloadHookFunc(func(_ node.Node) {
		count += 1
	})

	hooks.AddUnloadHook(h)

	hooks.Unload(n)
	assert.Equal(t, 1, count)
}
