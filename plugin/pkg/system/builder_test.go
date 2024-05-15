package system

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/event"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestAddToHook(t *testing.T) {
	h := hook.New()

	b := event.NewBroker()
	defer b.Close()

	err := AddToHook(Config{
		Broker: b,
	})(h)
	assert.NoError(t, err)

	n := node.NewManyToOneNode(nil)
	defer n.Close()

	sym := symbol.New(&scheme.SpecMeta{}, n)

	err = h.Load(sym)
	assert.NoError(t, err)

	err = h.Unload(sym)
	assert.NoError(t, err)
}

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	err := AddToScheme(Config{})(s)
	assert.NoError(t, err)

	testCase := []string{KindNative, KindTrigger}

	for _, tc := range testCase {
		t.Run(tc, func(t *testing.T) {
			_, ok := s.KnownType(tc)
			assert.True(t, ok)

			_, ok = s.Codec(tc)
			assert.True(t, ok)
		})
	}
}
