package hook

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/chart"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestHook_LinkHook(t *testing.T) {
	hooks := New()

	count := 0
	h := chart.LinkFunc(func(_ *chart.Chart) error {
		count += 1
		return nil
	})

	hooks.AddLinkHook(h)

	err := hooks.Link(&chart.Chart{})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestHook_UnlinkHook(t *testing.T) {
	hooks := New()

	count := 0
	h := chart.UnlinkFunc(func(_ *chart.Chart) error {
		count += 1
		return nil
	})

	hooks.AddUnlinkHook(h)

	err := hooks.Unlink(&chart.Chart{})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestHook_LoadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(nil)

	count := 0
	h := symbol.LoadFunc(func(_ *symbol.Symbol) error {
		count += 1
		return nil
	})

	hooks.AddLoadHook(h)

	err := hooks.Load(&symbol.Symbol{
		Spec: &spec.Meta{},
		Node: n,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestHook_UnloadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(nil)

	count := 0
	h := symbol.UnloadFunc(func(_ *symbol.Symbol) error {
		count += 1
		return nil
	})

	hooks.AddUnloadHook(h)

	err := hooks.Unload(&symbol.Symbol{
		Spec: &spec.Meta{},
		Node: n,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
