package network

import (
	"fmt"
	"testing"

	"github.com/phayes/freeport"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/require"
)

func TestAddToHook(t *testing.T) {
	h := hook.New()

	err := AddToHook().AddToHook(h)
	require.NoError(t, err)

	port, err := freeport.GetFreePort()
	require.NoError(t, err)

	n := NewHTTPListenNode(fmt.Sprintf(":%d", port))
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

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	err := AddToScheme().AddToScheme(s)
	require.NoError(t, err)

	tests := []string{KindHTTP, KindListener, KindRouter, KindWebSocket, KindUpgrader}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			require.NotNil(t, s.KnownType(tt))
			require.NotNil(t, s.Codec(tt))
		})
	}
}
