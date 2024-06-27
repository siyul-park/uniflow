package control

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/x/pkg/language"
	"github.com/siyul-park/uniflow/x/pkg/language/expr"
	"github.com/stretchr/testify/assert"
)

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	m := language.NewModule()
	m.Store(expr.Kind, expr.NewCompiler())

	err := AddToScheme(m, expr.Kind)(s)
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
