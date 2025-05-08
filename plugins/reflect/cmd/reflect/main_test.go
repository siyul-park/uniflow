package main

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/stretchr/testify/require"
)

func TestPlugin_Load(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	hb := hook.NewBuilder()

	agent := runtime.NewAgent()
	defer agent.Close()

	p.SetHookBuilder(hb)
	p.SetAgent(agent)

	err := p.Load(ctx)
	require.NoError(t, err)
}

func TestPlugin_Unload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	hb := hook.NewBuilder()

	agent := runtime.NewAgent()
	defer agent.Close()

	p.SetHookBuilder(hb)
	p.SetAgent(agent)

	err := p.Load(ctx)
	require.NoError(t, err)

	err = p.Unload(ctx)
	require.NoError(t, err)
}
