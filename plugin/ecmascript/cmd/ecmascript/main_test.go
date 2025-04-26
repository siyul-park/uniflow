package main

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/plugin/ecmascript/pkg/javascript"
	"github.com/siyul-park/uniflow/plugin/ecmascript/pkg/typescript"
)

func TestPlugin_Load(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	r := language.NewRegistry()
	p.SetRegistry(r)

	err := p.Load(ctx)
	require.NoError(t, err)

	c, err := r.Lookup(javascript.Language)
	require.NoError(t, err)
	require.NotNil(t, c)

	c, err = r.Lookup(typescript.Language)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestPlugin_Unload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	r := language.NewRegistry()
	p.SetRegistry(r)

	err := p.Load(ctx)
	require.NoError(t, err)

	err = p.Unload(ctx)
	require.NoError(t, err)
}
