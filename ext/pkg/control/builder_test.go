package control

import (
	"testing"

	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/ext/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestAddToHook(t *testing.T) {
	h := hook.New()

	err := AddToHook()(h)
	assert.NoError(t, err)

	n := NewBlockNode(node.NewOneToOneNode(nil))
	defer n.Close()

	sym := symbol.New(&spec.Meta{}, n)

	err = h.Load(sym)
	assert.NoError(t, err)

	err = h.Unload(sym)
	assert.NoError(t, err)
}

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	m := language.NewModule()
	m.Store(text.Language, text.NewCompiler())

	err := AddToScheme(m, text.Language)(s)
	assert.NoError(t, err)

	testCase := []string{KindCall, KindFork, KindIf, KindLoop, KindMerge, KindNOP, KindSnippet, KindSwitch}

	for _, tc := range testCase {
		t.Run(tc, func(t *testing.T) {
			_, ok := s.KnownType(tc)
			assert.True(t, ok)

			_, ok = s.Codec(tc)
			assert.True(t, ok)
		})
	}
}
