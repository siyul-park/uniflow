package encoding

import (
	"github.com/go-faker/faker/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"unsafe"
)

func TestAssembler_Add(t *testing.T) {
	a := NewAssembler[any, any]()
	a.Add(CompilerFunc[any](func(typ reflect.Type) (Encoder[any, unsafe.Pointer], error) {
		return nil, nil
	}))

	assert.Equal(t, 1, a.Len())
}

func TestAssembler_Compile(t *testing.T) {
	a := NewAssembler[any, any]()
	a.Add(CompilerFunc[any](func(typ reflect.Type) (Encoder[any, unsafe.Pointer], error) {
		if typ.Elem().Kind() == reflect.String {
			return EncodeFunc[any, unsafe.Pointer](func(source any, target unsafe.Pointer) error {
				return nil
			}), nil
		}
		return nil, errors.WithStack(ErrUnsupportedValue)
	}))

	d, err := a.Compile(reflect.TypeOf(""))
	assert.NoError(t, err)
	assert.NotNil(t, d)
}

func TestAssembler_Decode(t *testing.T) {
	a := NewAssembler[any, any]()
	a.Add(CompilerFunc[any](func(typ reflect.Type) (Encoder[any, unsafe.Pointer], error) {
		if typ.Elem().Kind() == reflect.String {
			return EncodeFunc[any, unsafe.Pointer](func(source any, target unsafe.Pointer) error {
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

	err := a.Encode(source, &target)
	assert.NoError(t, err)
	assert.Equal(t, source, target)
}
