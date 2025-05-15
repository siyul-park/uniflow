package main

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/stretchr/testify/require"
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

	dr := driver.NewRegistry()
	p.SetLanguageRegistry(dr)

	err := p.Load(ctx)
	require.NoError(t, err)

	c, err := dr.Lookup("mongodb")
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestPlugin_Unload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	dr := driver.NewRegistry()
	p.SetLanguageRegistry(dr)

	err := p.Load(ctx)
	require.NoError(t, err)

	err = p.Unload(ctx)
	require.NoError(t, err)
}
