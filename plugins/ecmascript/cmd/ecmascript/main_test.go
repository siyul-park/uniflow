package main

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/language"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/plugins/ecmascript/pkg/javascript"
	"github.com/siyul-park/uniflow/plugins/ecmascript/pkg/typescript"
)

func TestPlugin_Load(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	lr := language.NewRegistry()
	p.SetLanguageRegistry(lr)

	err := p.Load(ctx)
	require.NoError(t, err)

	c, err := lr.Lookup(javascript.Language)
	require.NoError(t, err)
	require.NotNil(t, c)

	c, err = lr.Lookup(typescript.Language)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestPlugin_Unload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	lr := language.NewRegistry()
	p.SetLanguageRegistry(lr)

	err := p.Load(ctx)
	require.NoError(t, err)

	err = p.Unload(ctx)
	require.NoError(t, err)
}
