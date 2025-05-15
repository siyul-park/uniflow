package main

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	testing2 "github.com/siyul-park/uniflow/pkg/testing"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/plugins/testing/pkg/node"
)

func TestPlugin_Name(t *testing.T) {
	p := New()
	require.Equal(t, name, p.Name())
}

func TestPlugin_Version(t *testing.T) {
	p := New()
	require.Equal(t, version, p.Version())
}

func TestPlugin_Load(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	hb := hook.NewBuilder()
	sb := scheme.NewBuilder()
	tr := testing2.NewRunner()

	p.SetHookBuilder(hb)
	p.SetSchemeBuilder(sb)
	p.SetTestingRunner(tr)

	err := p.Load(ctx)
	require.NoError(t, err)

	h, err := hb.Build()
	require.NoError(t, err)
	require.NotNil(t, h)

	s, err := sb.Build()
	require.NoError(t, err)

	tests := []string{
		node.KindTest,
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

	hb := hook.NewBuilder()
	sb := scheme.NewBuilder()
	tr := testing2.NewRunner()

	p.SetHookBuilder(hb)
	p.SetSchemeBuilder(sb)
	p.SetTestingRunner(tr)

	err := p.Load(ctx)
	require.NoError(t, err)

	err = p.Unload(ctx)
	require.NoError(t, err)
}
