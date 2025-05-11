package main

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/driver"
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

	drv := driver.New()
	defer drv.Close()

	conn, err := drv.Open(faker.UUIDHyphenated())
	require.NoError(t, err)

	p.SetHookBuilder(hb)
	p.SetAgent(agent)
	p.SetConn(conn)

	err = p.Load(ctx)
	require.NoError(t, err)

	h, err := hb.Build()
	require.NoError(t, err)
	require.NotNil(t, h)
}

func TestPlugin_Unload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := New()

	hb := hook.NewBuilder()

	agent := runtime.NewAgent()
	defer agent.Close()

	drv := driver.New()
	defer drv.Close()

	conn, err := drv.Open(faker.UUIDHyphenated())
	require.NoError(t, err)

	p.SetHookBuilder(hb)
	p.SetAgent(agent)
	p.SetConn(conn)

	err = p.Load(ctx)
	require.NoError(t, err)

	err = p.Unload(ctx)
	require.NoError(t, err)
}
