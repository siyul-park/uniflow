package types

import (
	"errors"
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/stretchr/testify/assert"
)

func TestError_NewError(t *testing.T) {
	err := errors.New("test error")
	e := NewError(err)

	assert.NotNil(t, e)
	assert.Equal(t, err, e.value)
}

func TestError_Error(t *testing.T) {
	err := errors.New("test error")
	e := NewError(err)

	assert.Equal(t, "test error", e.Error())
}

func TestError_Kind(t *testing.T) {
	err := errors.New("test error")
	e := NewError(err)

	assert.Equal(t, KindError, e.Kind())
}

func TestError_Hash(t *testing.T) {
	err1 := errors.New("test error 1")
	err2 := errors.New("test error 2")

	e1 := NewError(err1)
	e2 := NewError(err2)

	assert.NotEqual(t, e1.Hash(), e2.Hash())
}

func TestError_Interface(t *testing.T) {
	err := errors.New("test error")
	e := NewError(err)

	assert.Equal(t, err, e.Interface())
}

func TestError_Equal(t *testing.T) {
	err1 := errors.New("test error 1")
	err2 := errors.New("test error 2")

	e1 := NewError(err1)
	e2 := NewError(err2)

	assert.True(t, e1.Equal(e1))
	assert.True(t, e2.Equal(e2))
	assert.False(t, e1.Equal(e2))
}

func TestError_Compare(t *testing.T) {
	err1 := errors.New("test error 1")
	err2 := errors.New("test error 2")

	e1 := NewError(err1)
	e2 := NewError(err2)

	assert.Equal(t, 0, e1.Compare(e1))
	assert.NotEqual(t, 0, e1.Compare(e2))
	assert.NotEqual(t, 0, e2.Compare(e1))
}

func TestError_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Object]()
	enc.Add(newErrorEncoder())

	err := errors.New("test error")
	e := NewError(err)

	encoded, encodeErr := enc.Encode(err)
	assert.NoError(t, encodeErr)
	assert.Equal(t, e, encoded)
}

func TestError_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Object, any]()
	dec.Add(newErrorDecoder())

	err := errors.New("test error")
	e := NewError(err)

	var decoded error
	assert.NoError(t, dec.Decode(e, &decoded))
	assert.Equal(t, err, decoded)
}
