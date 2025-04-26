package plugin

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	m := NewMock(t)

	err := r.Register(m)
	require.NoError(t, err)
}

func TestRegistry_Inject(t *testing.T) {
	r := NewRegistry()
	m := NewMock(t)

	err := r.Register(m)
	require.NoError(t, err)

	m.On("SetXXX", "foo").Return(nil)

	err = r.Inject("foo")
	require.NoError(t, err)
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
