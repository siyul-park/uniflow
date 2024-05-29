package object

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestNewBool(t *testing.T) {
	v := NewBool(true)

	assert.Equal(t, KindBool, v.Kind())
	assert.Equal(t, true, v.Interface())
	assert.Equal(t, true, v.Bool())
}

func TestBool_Compare(t *testing.T) {
	assert.Equal(t, 0, TRUE.Compare(TRUE))
	assert.Equal(t, 0, FALSE.Compare(FALSE))
	assert.Equal(t, 1, TRUE.Compare(FALSE))
	assert.Equal(t, -1, FALSE.Compare(TRUE))
	assert.Equal(t, 1, TRUE.Compare(FALSE))
	assert.Equal(t, -1, FALSE.Compare(TRUE))
}

func TestBool_Encode(t *testing.T) {
	enc := encoding.NewAssembler[*Object, any]()
	enc.Add(newBoolEncoder())

	source := true
	v := NewBool(source)

	var decoded Object
	err := enc.Encode(&decoded, &source)
	assert.NoError(t, err)
	assert.Equal(t, v, decoded)
}

func TestBool_Decode(t *testing.T) {
	dec := encoding.NewAssembler[Object, any]()
	dec.Add(newBoolDecoder())

	t.Run("bool", func(t *testing.T) {
		v := NewBool(true)

		var decoded bool
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, true, decoded)
	})

	t.Run("any", func(t *testing.T) {
		v := NewBool(true)

		var decoded any
		err := dec.Encode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, true, decoded)
	})
}
