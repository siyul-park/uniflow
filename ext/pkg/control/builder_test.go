package control

import (
	"testing"

	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/ext/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/require"
)

func TestAddToScheme(t *testing.T) {
	s := scheme.New()

	err := AddToScheme(map[string]language.Compiler{text.Language: text.NewCompiler()}, text.Language).AddToScheme(s)
	require.NoError(t, err)

	tests := []string{KindBlock, KindCache, KindFor, KindFork, KindIf, KindFor, KindMerge, KindNOP, KindPipe, KindRetry, KindStep, KindSession, KindSleep, KindSnippet, KindSplit, KindSwitch, KindThrow, KindTry}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			require.NotNil(t, s.KnownType(tt))
			require.NotNil(t, s.Codec(tt))
		})
	}
}
