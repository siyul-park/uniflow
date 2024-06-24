package control

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/event"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/plugin/pkg/language"
	"github.com/siyul-park/uniflow/plugin/pkg/language/expr"
	"github.com/stretchr/testify/assert"
)

func TestAddToHook(t *testing.T) {
	h := hook.New()

	b := event.NewBroker()
	defer b.Close()

	m := language.NewModule()
	m.Store(expr.Kind, expr.NewCompiler())

	err := AddToHook(Config{
		Broker:     b,
		Module:     m,
		Expression: expr.Kind,
	})(h)
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

	m := language.NewModule()
	m.Store(expr.Kind, expr.NewCompiler())

	err := AddToScheme(Config{
		Broker:     b,
		Module:     m,
		Expression: expr.Kind,
	})(s)
	assert.NoError(t, err)

	testCase := []string{KindCall, KindIf, KindLoop, KindMerge, KindNOP, KindSnippet, KindSwitch}

	for _, tc := range testCase {
		t.Run(tc, func(t *testing.T) {
			_, ok := s.KnownType(tc)
			assert.True(t, ok)

			_, ok = s.Codec(tc)
			assert.True(t, ok)
		})
	}
}
