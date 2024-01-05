package process

import (
	"context"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

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
	assert.Equal(t, k2, h1)

	h2, ok := st.Pop(k3, v2)
	assert.True(t, ok)
	assert.Equal(t, k2, h2)

	h3, ok := st.Pop(k3, v1)
	assert.True(t, ok)
	assert.Equal(t, k1, h3)

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

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			st.Wait()
			close(done)
		}()

		select {
		case <-ctx.Done():
			assert.Fail(t, "timeout")
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

func BenchmarkStack_Push(b *testing.B) {
	st := newStack()
	defer st.Close()

	for i := 0; i < b.N; i++ {
		k := ulid.Make()
		v := ulid.Make()

		st.Push(k, v)
	}
}

func BenchmarkStack_Pop(b *testing.B) {
	st := newStack()
	defer st.Close()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		k := ulid.Make()
		v := ulid.Make()

		st.Push(k, v)

		b.StartTimer()

		st.Pop(k, v)
	}
}

func BenchmarkStack_Clear(b *testing.B) {
	st := newStack()
	defer st.Close()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		k := ulid.Make()
		v := ulid.Make()

		st.Push(k, v)

		b.StartTimer()

		st.Clear(k)
	}
}

func BenchmarkStack_Len(b *testing.B) {
	st := newStack()
	defer st.Close()

	k := ulid.Make()
	v := ulid.Make()

	st.Push(k, v)

	for i := 0; i < b.N; i++ {
		st.Len(k)
	}
}
