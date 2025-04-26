package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProxy_Inject(t *testing.T) {
	m := NewMock(t)
	p := NewProxy(m)

	m.On("SetXXX", "foo").Return(nil)

	err := p.Inject("foo")
	require.NoError(t, err)
}

func TestProxy_Load(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	m := NewMock(t)
	p := NewProxy(m)

	m.On("Load", ctx).Return(nil)

	err := p.Load(ctx)
	require.NoError(t, err)
}

func TestProxy_Unload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	m := NewMock(t)
	p := NewProxy(m)

	m.On("Unload", ctx).Return(nil)

	err := p.Unload(ctx)
	require.NoError(t, err)
}

func TestProxy_Unwrap(t *testing.T) {
	m := NewMock(t)
	p := NewProxy(m)

	u := p.Unwrap()
	require.Equal(t, m, u)
}
