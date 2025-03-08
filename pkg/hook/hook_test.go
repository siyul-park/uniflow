package hook

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/require"
)

func TestHook_LoadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(nil)

	count := 0
	h := symbol.LoadFunc(func(_ *symbol.Symbol) error {
		count += 1
		return nil
	})

	require.True(t, hooks.AddLoadHook(h))
	require.False(t, hooks.AddLoadHook(h))

	err := hooks.Load(&symbol.Symbol{
		Spec: &spec.Meta{},
		Node: n,
	})
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestHook_UnloadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(nil)

	count := 0
	h := symbol.UnloadFunc(func(_ *symbol.Symbol) error {
		count += 1
		return nil
	})

	require.True(t, hooks.AddUnloadHook(h))
	require.False(t, hooks.AddUnloadHook(h))

	err := hooks.Unload(&symbol.Symbol{
		Spec: &spec.Meta{},
		Node: n,
	})
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
