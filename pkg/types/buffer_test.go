package types

import (
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/stretchr/testify/assert"
)

func TestBuffer_Read(t *testing.T) {
	r := strings.NewReader("test")
	b := NewBuffer(r)
	p := make([]byte, 4)
	n, err := b.Read(p)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "test", string(p))
}

func TestBuffer_Bytes(t *testing.T) {
	r := strings.NewReader("test")
	b := NewBuffer(r)
	p, err := b.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, "test", string(p))
}

func TestBuffer_Close(t *testing.T) {
	r := strings.NewReader("test")
	b := NewBuffer(r)
	assert.NoError(t, b.Close())
}

func TestBuffer_Kind(t *testing.T) {
	r := strings.NewReader("test")
	b := NewBuffer(r)
	assert.Equal(t, KindBuffer, b.Kind())
}

func TestBuffer_Hash(t *testing.T) {
	r1 := strings.NewReader("test1")
	r2 := strings.NewReader("test2")
	b1 := NewBuffer(r1)
	b2 := NewBuffer(r2)
	assert.NotEqual(t, b1.Hash(), b2.Hash())
}

func TestBuffer_Interface(t *testing.T) {
	r := strings.NewReader("test")
	b := NewBuffer(r)
	assert.Equal(t, r, b.Interface())
}

func TestBuffer_Equal(t *testing.T) {
	r1 := strings.NewReader("test1")
	r2 := strings.NewReader("test2")
	b1 := NewBuffer(r1)
	b2 := NewBuffer(r2)
	assert.True(t, b1.Equal(b1))
	assert.False(t, b1.Equal(b2))
}

func TestBuffer_Compare(t *testing.T) {
	r1 := strings.NewReader("test1")
	r2 := strings.NewReader("test2")
	b1 := NewBuffer(r1)
	b2 := NewBuffer(r2)
	assert.Equal(t, 0, b1.Compare(b1))
	assert.NotEqual(t, 0, b1.Compare(b2))
}

func TestBuffer_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newBufferEncoder())

	t.Run("io.Reader", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		decoded, err := enc.Encode(source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestBuffer_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newBufferDecoder())

	t.Run("io.Reader", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		var decoded io.Reader
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("slice", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		var decoded []byte
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, []byte("test"), decoded)
	})

	t.Run("array", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		var decoded [3]byte
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, []byte("test"), decoded)
	})

	t.Run("string", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		var decoded string
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)

		d, err := base64.StdEncoding.DecodeString(decoded)
		assert.NoError(t, err)
		assert.Equal(t, []byte("test"), d)
	})

	t.Run("any", func(t *testing.T) {
		source := strings.NewReader("test")
		v := NewBuffer(source)

		var decoded any
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})
}
