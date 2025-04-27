package main

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/plugins/cel/pkg/cel"
)

func TestPlugin_Load(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New(Config{})

	lr := language.NewRegistry()
	p.SetLanguageRegistry(lr)

	err := p.Load(ctx)
	require.NoError(t, err)

	c, err := lr.Lookup(cel.Language)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestPlugin_Unload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New(Config{})

	lr := language.NewRegistry()
	p.SetLanguageRegistry(lr)

	err := p.Load(ctx)
	require.NoError(t, err)

	err = p.Unload(ctx)
	require.NoError(t, err)
}
