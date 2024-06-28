package control

import (
	"testing"

	"github.com/siyul-park/uniflow/ext/language"
	"github.com/siyul-park/uniflow/ext/language/text"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/stretchr/testify/assert"
)

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	m := language.NewModule()
	m.Store(text.Kind, text.NewCompiler())

	err := AddToScheme(m, text.Kind)(s)
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
