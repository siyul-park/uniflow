package primitive

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestNewBuffer(t *testing.T) {
	v := NewBuffer(bytes.NewBuffer([]byte{1}))

	assert.Equal(t, KindBuffer, v.Kind())
	assert.Equal(t, []byte{1}, v.Bytes())
}

func TestBuffer_Compare(t *testing.T) {
	v1 := NewBuffer(bytes.NewBuffer([]byte{0}))
	v2 := NewBuffer(bytes.NewBuffer([]byte{1}))

	assert.Equal(t, 0, v1.Compare(v1))
	assert.Equal(t, -1, v1.Compare(v2))
	assert.Equal(t, 1, v2.Compare(v1))
}

func TestBuffer_EncodeAndDecode(t *testing.T) {
	e := newBufferEncoder()
	d := newBufferDecoder()

	t.Run("io.Reader", func(t *testing.T) {
		source := bytes.NewBuffer([]byte{1})

		encoded, err := e.Encode(source)
		assert.NoError(t, err)

		var decoded io.Reader
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)

		data, err := io.ReadAll(decoded)
		assert.NoError(t, err)
		assert.Equal(t, []byte{1}, data)
	})

	t.Run("[]bytes", func(t *testing.T) {
		source := bytes.NewBuffer([]byte{1})

		encoded, err := e.Encode(source)
		assert.NoError(t, err)

		var decoded []byte
		err = d.Decode(encoded, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, []byte{1}, decoded)
	})
}
