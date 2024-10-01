package network

import (
	"fmt"
	"testing"

	"github.com/phayes/freeport"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestAddToHook(t *testing.T) {
	h := hook.New()

	err := AddToHook().AddToHooks(h)
	assert.NoError(t, err)

	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPListenNode(fmt.Sprintf(":%d", port))
	defer n.Close()

	sb := &symbol.Symbol{
		Spec: &spec.Meta{},
		Node: n,
	}

	err = h.Load(sb)
	assert.NoError(t, err)

	err = h.Unload(sb)
	assert.NoError(t, err)
}

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	err := AddToScheme().AddToScheme(s)
	assert.NoError(t, err)

	tests := []string{KindHTTP, KindListener, KindProxy, KindRouter, KindWebSocket, KindGateway}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			assert.NotNil(t, s.KnownType(tt))
			assert.NotNil(t, s.Codec(tt))
		})
	}
}
