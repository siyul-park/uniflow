package main

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/plugins/sql/pkg/node"
)

func TestPlugin_Load(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	sb := scheme.NewBuilder()

	p.SetSchemeBuilder(sb)

	err := p.Load(ctx)
	require.NoError(t, err)

	s, err := sb.Build()
	require.NoError(t, err)

	tests := []string{
		node.KindSQL,
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

	p.SetSchemeBuilder(sb)

	err := p.Load(ctx)
	require.NoError(t, err)

	err = p.Unload(ctx)
	require.NoError(t, err)
}
