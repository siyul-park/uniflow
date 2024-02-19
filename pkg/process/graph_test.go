package process

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGraph_Add(t *testing.T) {
	g := newGraph()

	v1 := uuid.Must(uuid.NewV7())
	v2 := uuid.Must(uuid.NewV7())

	g.Add(v1, v2)

	assert.True(t, g.Has(v1, v2))
	assert.False(t, g.Has(v2, v1))

	g.Add(v1, v2)

	assert.True(t, g.Has(v1, v2))
	assert.False(t, g.Has(v2, v1))
}

func TestGraph_Delete(t *testing.T) {
	g := newGraph()

	v1 := uuid.Must(uuid.NewV7())
	v2 := uuid.Must(uuid.NewV7())

	g.Add(v1, v2)

	g.Delete(v1, v2)

	assert.False(t, g.Has(v1, v2))
	assert.False(t, g.Has(v2, v1))

	g.Delete(v1, v2)

	assert.False(t, g.Has(v1, v2))
	assert.False(t, g.Has(v2, v1))
}

func TestGraph_Upwards(t *testing.T) {
	g := newGraph()

	v1 := uuid.Must(uuid.NewV7())
	v2 := uuid.Must(uuid.NewV7())

	g.Add(v1, v2)

	var trace []uuid.UUID
	g.Upwards(v2, func(v uuid.UUID) bool {
		trace = append(trace, v)
		return true
	})
	assert.Equal(t, []uuid.UUID{v2, v1}, trace)
}

func TestGraph_Downwards(t *testing.T) {
	g := newGraph()

	v1 := uuid.Must(uuid.NewV7())
	v2 := uuid.Must(uuid.NewV7())

	g.Add(v1, v2)

	var trace []uuid.UUID
	g.Downwards(v1, func(v uuid.UUID) bool {
		trace = append(trace, v)
		return true
	})
	assert.Equal(t, []uuid.UUID{v1, v2}, trace)
}

func TestGraph_Close(t *testing.T) {
	g := newGraph()

	v1 := uuid.Must(uuid.NewV7())
	v2 := uuid.Must(uuid.NewV7())

	g.Add(v1, v2)

	g.Close()
	assert.False(t, g.Has(v1, v2))
	assert.False(t, g.Has(v2, v1))
}
