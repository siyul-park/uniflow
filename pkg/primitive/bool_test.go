package primitive

import (
	"github.com/siyul-park/uniflow/pkg/encoding"
	"testing"

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
	enc := encoding.NewCompiledDecoder[*Value, any]()
	enc.Add(newBoolEncoder())

	source := true
	binary := NewBool(true)

	var decoded Value
	err := enc.Decode(&decoded, &source)
	assert.NoError(t, err)
	assert.Equal(t, binary, decoded)
}

func TestBool_Decode(t *testing.T) {
	dec := encoding.NewCompiledDecoder[Value, any]()
	dec.Add(newBoolDecoder())

	t.Run("bool", func(t *testing.T) {
		v := NewBool(true)

		var decoded bool
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, true, decoded)
	})

	t.Run("any", func(t *testing.T) {
		v := NewBool(true)

		var decoded any
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, true, decoded)
	})
}

func BenchmarkBool_Encode(b *testing.B) {
	enc := encoding.NewCompiledDecoder[*Value, any]()
	enc.Add(newBoolEncoder())

	source := true

	for i := 0; i < b.N; i++ {
		var decoded Value
		_ = enc.Decode(&decoded, &source)
	}
}

func BenchmarkBool_Decode(b *testing.B) {
	dec := encoding.NewCompiledDecoder[Value, any]()
	dec.Add(newBoolDecoder())

	b.Run("bool", func(b *testing.B) {
		v := NewBool(true)

		for i := 0; i < b.N; i++ {
			var decoded bool
			_ = dec.Decode(v, &decoded)
		}
	})

	b.Run("any", func(b *testing.B) {
		v := NewBool(true)

		for i := 0; i < b.N; i++ {
			var decoded any
			_ = dec.Decode(v, &decoded)
		}
	})
}
