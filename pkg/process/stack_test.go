package process

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStack_Push(t *testing.T) {
	t.Run("Flat", func(t *testing.T) {
		g := newGraph()
		s := newStack(g)

		k := uuid.Must(uuid.NewV7())
		v := uuid.Must(uuid.NewV7())

		s.Push(k, v)
		assert.Equal(t, 1, s.Size(k))
	})

	t.Run("Deep", func(t *testing.T) {
		g := newGraph()
		s := newStack(g)

		k1 := uuid.Must(uuid.NewV7())
		k2 := uuid.Must(uuid.NewV7())

		v1 := uuid.Must(uuid.NewV7())
		v2 := uuid.Must(uuid.NewV7())

		g.Add(k1, k2)

		s.Push(k1, v1)
		s.Push(k2, v2)
		assert.Equal(t, 2, s.Size(k2))
	})

	t.Run("Recursive", func(t *testing.T) {
		g := newGraph()
		s := newStack(g)

		k1 := uuid.Must(uuid.NewV7())
		k2 := uuid.Must(uuid.NewV7())

		v1 := uuid.Must(uuid.NewV7())
		v2 := uuid.Must(uuid.NewV7())

		g.Add(k1, k2)
		g.Add(k2, k1)

		s.Push(k1, v1)
		s.Push(k2, v2)
		assert.Equal(t, 2, s.Size(k2))
	})
}

func TestStack_Pop(t *testing.T) {
	t.Run("Flat", func(t *testing.T) {
		g := newGraph()
		s := newStack(g)

		k := uuid.Must(uuid.NewV7())
		v := uuid.Must(uuid.NewV7())

		s.Push(k, v)

		head, ok := s.Pop(k, v)
		assert.True(t, ok)
		assert.Equal(t, k, head)
		assert.Equal(t, 0, s.Size(k))
	})

	t.Run("Deep", func(t *testing.T) {
		g := newGraph()
		s := newStack(g)

		k1 := uuid.Must(uuid.NewV7())
		k2 := uuid.Must(uuid.NewV7())

		v1 := uuid.Must(uuid.NewV7())
		v2 := uuid.Must(uuid.NewV7())

		g.Add(k1, k2)

		s.Push(k1, v1)
		s.Push(k2, v2)

		head, ok := s.Pop(k2, v2)
		assert.True(t, ok)
		assert.Equal(t, k2, head)
		assert.Equal(t, 1, s.Size(k2))

		head, ok = s.Pop(k2, v1)
		assert.True(t, ok)
		assert.Equal(t, k1, head)
		assert.Equal(t, 0, s.Size(k2))
	})

	t.Run("Recursive", func(t *testing.T) {
		g := newGraph()
		s := newStack(g)

		k1 := uuid.Must(uuid.NewV7())
		k2 := uuid.Must(uuid.NewV7())

		v1 := uuid.Must(uuid.NewV7())
		v2 := uuid.Must(uuid.NewV7())

		g.Add(k1, k2)
		g.Add(k2, k1)

		s.Push(k1, v1)
		s.Push(k2, v2)

		head, ok := s.Pop(k2, v2)
		assert.True(t, ok)
		assert.Equal(t, k2, head)
		assert.Equal(t, 1, s.Size(k2))

		head, ok = s.Pop(k2, v1)
		assert.True(t, ok)
		assert.Equal(t, k1, head)
		assert.Equal(t, 0, s.Size(k2))
	})
}

func TestStack_Clear(t *testing.T) {
	t.Run("Flat", func(t *testing.T) {
		g := newGraph()
		s := newStack(g)

		k := uuid.Must(uuid.NewV7())
		v := uuid.Must(uuid.NewV7())

		s.Push(k, v)

		s.Clear(k)
		assert.Equal(t, 0, s.Size(k))
	})

	t.Run("Deep", func(t *testing.T) {
		g := newGraph()
		s := newStack(g)

		k1 := uuid.Must(uuid.NewV7())
		k2 := uuid.Must(uuid.NewV7())

		v1 := uuid.Must(uuid.NewV7())
		v2 := uuid.Must(uuid.NewV7())

		g.Add(k1, k2)

		s.Push(k1, v1)
		s.Push(k2, v2)

		s.Clear(k2)
		assert.Equal(t, 0, s.Size(k2))
	})
}

func TestStack_Heads(t *testing.T) {
	t.Run("Flat", func(t *testing.T) {
		g := newGraph()
		s := newStack(g)

		k := uuid.Must(uuid.NewV7())
		v := uuid.Must(uuid.NewV7())

		s.Push(k, v)

		heads := s.Heads(k)
		assert.Equal(t, []uuid.UUID{k}, heads)
	})

	t.Run("Deep", func(t *testing.T) {
		g := newGraph()
		s := newStack(g)

		k1 := uuid.Must(uuid.NewV7())
		k2 := uuid.Must(uuid.NewV7())

		v1 := uuid.Must(uuid.NewV7())

		g.Add(k1, k2)

		s.Push(k1, v1)

		heads := s.Heads(k2)
		assert.Equal(t, []uuid.UUID{k1}, heads)
	})
}
