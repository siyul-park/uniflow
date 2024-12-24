package packet

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	pck := New(nil)
	assert.NotNil(t, pck)
}

func TestJoin(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		res := Join(None, None)
		assert.Equal(t, None, res)
	})

	t.Run("Zero", func(t *testing.T) {
		res := Join()
		assert.Equal(t, None, res)
	})

	t.Run("One", func(t *testing.T) {
		pck := New(nil)
		res := Join(pck)
		assert.Equal(t, pck, res)
	})

	t.Run("Many", func(t *testing.T) {
		pck1 := New(nil)
		pck2 := New(nil)
		res := Join(pck1, pck2)
		assert.Equal(t, types.NewSlice(nil, nil), res.Payload())
	})
}
