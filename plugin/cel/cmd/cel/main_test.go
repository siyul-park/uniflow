package main

import (
	"context"
	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/siyul-park/uniflow/plugin/cel/pkg/cel"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPlugin_Load(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New(Config{})

	r := language.NewRegistry()
	p.SetRegistry(r)

	err := p.Load(ctx)
	require.NoError(t, err)

	c, err := r.Lookup(cel.Language)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestPlugin_Unload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New(Config{})

	r := language.NewRegistry()
	p.SetRegistry(r)

	err := p.Load(ctx)
	require.NoError(t, err)

	err = p.Unload(ctx)
	require.NoError(t, err)
}
