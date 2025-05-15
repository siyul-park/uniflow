package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	m := NewMock(t)

	err := r.Register(m)
	require.NoError(t, err)
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()
	m := NewMock(t)

	err := r.Register(m)
	require.NoError(t, err)

	err = r.Unregister(m)
	require.NoError(t, err)
}

func TestRegistry_Inject(t *testing.T) {
	r := NewRegistry()
	m := NewMock(t)

	err := r.Register(m)
	require.NoError(t, err)

	m.On("SetXXX", "foo").Return(nil)

	count, err := r.Inject("foo")
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestRegistry_Plugins(t *testing.T) {
	r := NewRegistry()
	m1 := NewMock(t)
	m2 := NewMock(t)

	err := r.Register(m1)
	require.NoError(t, err)
	err = r.Register(m2)
	require.NoError(t, err)

	plugins := r.Plugins()
	require.Len(t, plugins, 2)
	require.Contains(t, plugins, m1)
	require.Contains(t, plugins, m2)
}

func TestRegistry_Load(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	r := NewRegistry()
	m := NewMock(t)

	err := r.Register(m)
	require.NoError(t, err)

	m.On("Load", ctx).Return(nil)

	err = r.Load(ctx)
	require.NoError(t, err)
}

func TestRegistry_Unload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	r := NewRegistry()
	m := NewMock(t)

	err := r.Register(m)
	require.NoError(t, err)

	m.On("Unload", ctx).Return(nil)

	err = r.Unload(ctx)
	require.NoError(t, err)
}
