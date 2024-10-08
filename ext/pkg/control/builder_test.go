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

	err := AddToHook().AddToHooks(h)
	assert.NoError(t, err)

	n := NewBlockNode(&symbol.Symbol{
		Node: node.NewOneToOneNode(nil),
	})
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

	m := language.NewModule()
	m.Store(text.Language, text.NewCompiler())

	err := AddToScheme(m, text.Language).AddToScheme(s)
	assert.NoError(t, err)

	tests := []string{KindCall, KindFork, KindIf, KindLoop, KindMerge, KindNOP, KindReduce, KindSession, KindSnippet, KindSplit, KindSwitch}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			assert.NotNil(t, s.KnownType(tt))
			assert.NotNil(t, s.Codec(tt))
		})
	}
}
