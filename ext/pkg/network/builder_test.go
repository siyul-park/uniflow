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

	sym := symbol.New(&spec.Meta{}, n)

	err = h.Load(sym)
	assert.NoError(t, err)

	err = h.Unload(sym)
	assert.NoError(t, err)
}

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	err := AddToScheme().AddToScheme(s)
	assert.NoError(t, err)

	testCase := []string{KindHTTP, KindListener, KindRouter, KindWebSocket, KindGateway}

	for _, tc := range testCase {
		t.Run(tc, func(t *testing.T) {
			_, ok := s.KnownType(tc)
			assert.True(t, ok)

			_, ok = s.Codec(tc)
			assert.True(t, ok)
		})
	}
}
