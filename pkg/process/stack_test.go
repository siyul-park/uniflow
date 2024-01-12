package process

import (
	"context"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestStack_Link(t *testing.T) {
	st := newStack()
	defer st.Close()

	k1 := ulid.Make()
	k2 := ulid.Make()

	st.Link(k1, k2)

	assert.Equal(t, []ulid.ULID{k1}, st.Stems(k2))
	assert.Equal(t, []ulid.ULID{k2}, st.Leaves(k1))

	st.Link(k1, k2)

	assert.Equal(t, []ulid.ULID{k1}, st.Stems(k2))
	assert.Equal(t, []ulid.ULID{k2}, st.Leaves(k1))
}

func TestStack_Unlink(t *testing.T) {
	st := newStack()
	defer st.Close()

	k1 := ulid.Make()
	k2 := ulid.Make()

	st.Link(k1, k2)

	st.Unlink(k1, k2)

	assert.Len(t, st.Stems(k2), 0)
	assert.Len(t, st.Leaves(k1), 0)

	st.Unlink(k1, k2)

	assert.Len(t, st.Stems(k2), 0)
	assert.Len(t, st.Leaves(k1), 0)
}

func TestStack_Pop(t *testing.T) {
	st := newStack()
	defer st.Close()

	k1 := ulid.Make()
	k2 := ulid.Make()
	k3 := ulid.Make()

	v1 := ulid.Make()
	v2 := ulid.Make()
	v3 := ulid.Make()

	st.Link(k1, k2)
	st.Link(k2, k3)

	st.Push(k1, v1)
	st.Push(k2, v2)
	st.Push(k2, v3)

	h1, ok := st.Pop(k3, v3)
	assert.True(t, ok)
	assert.Contains(t, h1, k2)

	h2, ok := st.Pop(k3, v2)
	assert.True(t, ok)
	assert.Contains(t, h2, k1)

	h3, ok := st.Pop(k3, v1)
	assert.True(t, ok)
	assert.Len(t, h3, 0)

	assert.Equal(t, 0, st.Len(k3))
}

func TestStack_Len(t *testing.T) {
	st := newStack()
	defer st.Close()

	k1 := ulid.Make()
	k2 := ulid.Make()

	v1 := ulid.Make()
	v2 := ulid.Make()
	v3 := ulid.Make()

	st.Link(k1, k2)

	st.Push(k1, v1)
	st.Push(k2, v2)
	st.Push(k2, v3)

	assert.Equal(t, 1, st.Len(k1))
	assert.Equal(t, 3, st.Len(k2))
}
func TestStack_Wait(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		st := newStack()
		defer st.Close()

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			st.Wait()
			close(done)
		}()

		select {
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		case <-done:
		}
	})

	t.Run("Not Empty", func(t *testing.T) {
		st := newStack()
		defer st.Close()

		k1 := ulid.Make()
		v1 := ulid.Make()

		st.Push(k1, v1)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()

		done := make(chan struct{})
		go func() {
			st.Wait()
			close(done)
		}()

		select {
		case <-ctx.Done():
		case <-done:
			assert.Fail(t, "timeout")
		}
	})
}

func TestStack_Clear(t *testing.T) {
	st := newStack()
	defer st.Close()

	k1 := ulid.Make()
	k2 := ulid.Make()
	k3 := ulid.Make()
	k4 := ulid.Make()

	v1 := ulid.Make()
	v2 := ulid.Make()
	v3 := ulid.Make()

	st.Link(k1, k2)
	st.Link(k2, k3)
	st.Link(k2, k4)

	st.Push(k1, v1)
	st.Push(k2, v2)
	st.Push(k2, v3)

	st.Clear(k4)

	_, ok := st.Pop(k4, v3)
	assert.False(t, ok)

	_, ok = st.Pop(k4, v2)
	assert.False(t, ok)

	_, ok = st.Pop(k3, v3)
	assert.True(t, ok)

	_, ok = st.Pop(k3, v2)
	assert.True(t, ok)

	_, ok = st.Pop(k3, v1)
	assert.True(t, ok)

	assert.Equal(t, 0, st.Len(k3))
}
