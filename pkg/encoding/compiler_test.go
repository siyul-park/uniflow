package encoding

import (
	"github.com/go-faker/faker/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"unsafe"
)

func TestCompiledDecoder_Add(t *testing.T) {
	c := NewCompiledDecoder[any, any]()
	c.Add(CompilerFunc[any](func(typ reflect.Type) (Decoder[any, unsafe.Pointer], error) {
		return nil, nil
	}))

	assert.Equal(t, 1, c.Len())
}

func TestCompiledDecoder_Compile(t *testing.T) {
	c := NewCompiledDecoder[any, any]()
	c.Add(CompilerFunc[any](func(typ reflect.Type) (Decoder[any, unsafe.Pointer], error) {
		if typ.Elem().Kind() == reflect.String {
			return DecoderFunc[any, unsafe.Pointer](func(source any, target unsafe.Pointer) error {
				return nil
			}), nil
		}
		return nil, errors.WithStack(ErrUnsupportedValue)
	}))

	d, err := c.Compile(reflect.TypeOf(""))
	assert.NoError(t, err)
	assert.NotNil(t, d)
}

func TestCompiledDecoder_Decode(t *testing.T) {
	c := NewCompiledDecoder[any, any]()
	c.Add(CompilerFunc[any](func(typ reflect.Type) (Decoder[any, unsafe.Pointer], error) {
		if typ.Elem().Kind() == reflect.String {
			return DecoderFunc[any, unsafe.Pointer](func(source any, target unsafe.Pointer) error {
				if s, ok := source.(string); ok {
					*(*string)(target) = s
					return nil
				}
				return errors.WithStack(ErrUnsupportedValue)
			}), nil
		}
		return nil, errors.WithStack(ErrUnsupportedValue)
	}))

	source := faker.UUIDHyphenated()
	target := ""

	err := c.Decode(source, &target)
	assert.NoError(t, err)
	assert.Equal(t, source, target)
}
