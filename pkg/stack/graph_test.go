package stack

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestGraph_Push(t *testing.T) {
	g := NewGraph[string]()
	defer g.Close()

	n1 := faker.UUIDHyphenated()
	n2 := faker.UUIDHyphenated()

	g.Push("", n1)
	assert.True(t, g.Has("", n1))

	g.Push(n1, n2)
	assert.True(t, g.Has(n1, n2))
}

func TestGraph_Pop(t *testing.T) {
	g := NewGraph[string]()
	defer g.Close()

	n1 := faker.UUIDHyphenated()
	n2 := faker.UUIDHyphenated()
	n3 := faker.UUIDHyphenated()

	g.Push("", n1)
	g.Push(n1, n2)
	g.Push(n2, n3)

	assert.True(t, g.Pop(n3, n3))
	assert.True(t, g.Pop(n3, n2))
	assert.True(t, g.Pop(n3, n1))

	assert.False(t, g.Has("", n1))
	assert.False(t, g.Has("", n2))
	assert.False(t, g.Has("", n3))
}

func TestGraph_Clear(t *testing.T) {
	g := NewGraph[string]()
	defer g.Close()

	n1 := faker.UUIDHyphenated()
	n2 := faker.UUIDHyphenated()
	n3 := faker.UUIDHyphenated()

	g.Push("", n1)
	g.Push(n1, n2)
	g.Push(n2, n3)

	g.Clear(n3)

	assert.False(t, g.Has("", n1))
	assert.False(t, g.Has("", n2))
	assert.False(t, g.Has("", n3))
}

func TestGraph_Done(t *testing.T) {
	g := NewGraph[string]()
	defer g.Close()

	n1 := faker.UUIDHyphenated()
	n2 := faker.UUIDHyphenated()

	g.Push(n1, n2)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	done := g.Done("")

	g.Clear(n2)

	select {
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	case <-done:
	}
}
