package networkx

import (
	"fmt"
	"testing"

	"github.com/phayes/freeport"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestAddToHooks(t *testing.T) {
	hk := hook.New()

	err := AddToHooks()(hk)
	assert.NoError(t, err)

	port, err := freeport.GetFreePort()
	assert.NoError(t, err)

	n := NewHTTPNode(HTTPNodeConfig{
		Address: fmt.Sprintf(":%d", port),
	})

	hk.Load(n)
	
	errChan := make(chan error)

	err = n.WaitForListen(errChan)

	assert.NoError(t, err)
	assert.NoError(t, n.Close())
}

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	err := AddToScheme()(s)
	assert.NoError(t, err)

	_, ok := s.Codec(KindHTTP)
	assert.True(t, ok)

	_, ok = s.KnownType(KindHTTP)
	assert.True(t, ok)
}
