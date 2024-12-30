package control

import (
	"testing"

	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/ext/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	m := language.NewModule()
	m.Store(text.Language, text.NewCompiler())

	err := AddToScheme(m, text.Language).AddToScheme(s)
	assert.NoError(t, err)

	tests := []string{KindBlock, KindCache, KindFor, KindFork, KindIf, KindFor, KindMerge, KindNOP, KindPipe, KindReduce, KindRetry, KindStep, KindSession, KindSleep, KindSnippet, KindSplit, KindSwitch, KindTry}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			assert.NotNil(t, s.KnownType(tt))
			assert.NotNil(t, s.Codec(tt))
		})
	}
}
