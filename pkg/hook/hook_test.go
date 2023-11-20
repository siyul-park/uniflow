package hook

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestHook_PreLoadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(node.OneToOneNodeConfig{})

	h := symbol.PreLoadHookFunc(func(_ node.Node) error {
		return nil
	})

	hooks.AddPreLoadHook(h)

	err := hooks.PreLoad(n)
	assert.NoError(t, err)
}

func TestHook_PostLoadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(node.OneToOneNodeConfig{})

	h := symbol.PostLoadHookFunc(func(_ node.Node) error {
		return nil
	})

	hooks.AddPostLoadHook(h)

	err := hooks.PostLoad(n)
	assert.NoError(t, err)
}

func TestHook_PreUnloadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(node.OneToOneNodeConfig{})

	h := symbol.PreUnloadHookFunc(func(_ node.Node) error {
		return nil
	})

	hooks.AddPreUnloadHook(h)

	err := hooks.PreUnload(n)
	assert.NoError(t, err)
}

func TestHook_PostUnloadHook(t *testing.T) {
	hooks := New()

	n := node.NewOneToOneNode(node.OneToOneNodeConfig{})

	h := symbol.PostUnloadHookFunc(func(_ node.Node) error {
		return nil
	})

	hooks.AddPostUnloadHook(h)

	err := hooks.PostUnload(n)
	assert.NoError(t, err)
}
