package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	pck := New(nil)
	assert.NotNil(t, pck)
}

func TestMerge(t *testing.T) {
	t.Run("EOF", func(t *testing.T) {
		res := Merge([]*Packet{EOF, EOF})
		assert.Equal(t, EOF, res)
	})

	t.Run("Zero", func(t *testing.T) {
		res := Merge([]*Packet{})
		assert.Equal(t, EOF, res)
	})

	t.Run("One", func(t *testing.T) {
		pck := New(nil)
		res := Merge([]*Packet{pck})
		assert.Equal(t, pck, res)
	})

	t.Run("Many", func(t *testing.T) {
		pck1 := New(nil)
		pck2 := New(nil)
		res := Merge([]*Packet{pck1, pck2})
		assert.Equal(t, []any{nil, nil}, res.Payload().Interface())
	})
}
