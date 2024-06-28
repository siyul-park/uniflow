package event

import (
	"testing"

	"github.com/siyul-park/uniflow/event"
	"github.com/siyul-park/uniflow/hook"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/spec"
	"github.com/siyul-park/uniflow/symbol"
	"github.com/stretchr/testify/assert"
)

func TestAddToHook(t *testing.T) {
	h := hook.New()

	b := event.NewBroker()
	defer b.Close()

	err := AddToHook(b)(h)
	assert.NoError(t, err)

	n := node.NewManyToOneNode(nil)
	defer n.Close()

	sym := symbol.New(&spec.Meta{}, n)

	err = h.Load(sym)
	assert.NoError(t, err)

	err = h.Unload(sym)
	assert.NoError(t, err)
}

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	b := event.NewBroker()
	defer b.Close()

	err := AddToScheme(b)(s)
	assert.NoError(t, err)

	testCase := []string{KindTrigger}

	for _, tc := range testCase {
		t.Run(tc, func(t *testing.T) {
			_, ok := s.KnownType(tc)
			assert.True(t, ok)

			_, ok = s.Codec(tc)
			assert.True(t, ok)
		})
	}
}
