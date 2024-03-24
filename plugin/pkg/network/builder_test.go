package network

import (
	"fmt"
	"testing"

	"github.com/phayes/freeport"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestAddToHook(t *testing.T) {
	h := hook.New()

	err := AddToHook()(h)
	assert.NoError(t, err)

	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPNode(fmt.Sprintf(":%d", port))
	defer n.Close()

	err = h.Load(n)
	assert.NoError(t, err)
}

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	err := AddToScheme()(s)
	assert.NoError(t, err)

	testCase := []string{KindCHTTP, KindHTTP, KindRoute, KindWebsocket}

	for _, tc := range testCase {
		t.Run(tc, func(t *testing.T) {
			_, ok := s.KnownType(tc)
			assert.True(t, ok)

			_, ok = s.Codec(tc)
			assert.True(t, ok)
		})
	}
}
