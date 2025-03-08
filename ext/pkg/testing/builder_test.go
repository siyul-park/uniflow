package testing

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	testing2 "github.com/siyul-park/uniflow/pkg/testing"
	"github.com/stretchr/testify/require"
)

func TestAddToHook(t *testing.T) {
	h := hook.New()

	runner := testing2.NewRunner()

	err := AddToHook(runner).AddToHook(h)
	require.NoError(t, err)

	n := NewTestNode()
	defer n.Close()

	sb := &symbol.Symbol{
		Spec: &spec.Meta{},
		Node: n,
	}

	err = h.Load(sb)
	require.NoError(t, err)

	err = h.Unload(sb)
	require.NoError(t, err)
}
