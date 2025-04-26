package main

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/plugin/ctrl/pkg/node"
)

func TestPlugin_Load(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	sb := scheme.NewBuilder()
	lr := language.NewRegistry()

	lr.SetDefault(text.Language)

	err := lr.Register(text.Language, text.NewCompiler())
	require.NoError(t, err)

	p.SetSchemeBuilder(sb)
	p.SetLanguageRegistry(lr)

	err = p.Load(ctx)
	require.NoError(t, err)

	s, err := sb.Build()
	require.NoError(t, err)

	tests := []string{
		node.KindBlock,
		node.KindFor,
		node.KindFork,
		node.KindIf,
		node.KindMerge,
		node.KindNOP,
		node.KindPipe,
		node.KindRetry,
		node.KindSleep,
		node.KindSnippet,
		node.KindSplit,
		node.KindStep,
		node.KindSwitch,
		node.KindThrow,
		node.KindTry,
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			require.NotNil(t, s.KnownType(tt))
			require.NotNil(t, s.Codec(tt))
		})
	}
}

func TestPlugin_Unload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	sb := scheme.NewBuilder()
	lr := language.NewRegistry()

	p.SetSchemeBuilder(sb)
	p.SetLanguageRegistry(lr)

	err := p.Load(ctx)
	require.NoError(t, err)

	err = p.Unload(ctx)
	require.NoError(t, err)
}
