package encoding

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/go-faker/faker/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestEncodeAssembler_Add(t *testing.T) {
	a := NewEncodeAssembler[any, any]()
	a.Add(EncodeCompilerFunc[any, any](func(typ reflect.Type) (Encoder[any, any], error) {
		return nil, nil
	}))

	assert.Equal(t, 1, a.Len())
}

func TestEncodeAssembler_Compile(t *testing.T) {
	a := NewEncodeAssembler[any, any]()
	a.Add(EncodeCompilerFunc[any, any](func(typ reflect.Type) (Encoder[any, any], error) {
		if typ.Kind() == reflect.String {
			return EncodeFunc[any, any](func(source any) (any, error) {
				return source, nil
			}), nil
		}
		return nil, errors.WithStack(ErrUnsupportedType)
	}))

	source := "test"
	e, err := a.Compile(reflect.TypeOf(source))
	assert.NoError(t, err)
	assert.NotNil(t, e)
}

func TestEncodeAssembler_Encode(t *testing.T) {
	a := NewEncodeAssembler[any, any]()
	a.Add(EncodeCompilerFunc[any, any](func(typ reflect.Type) (Encoder[any, any], error) {
		if typ.Kind() == reflect.String {
			return EncodeFunc[any, any](func(source any) (any, error) {
				return source, nil
			}), nil
		}
		return nil, errors.WithStack(ErrUnsupportedType)
	}))

	source := "test"
	target, err := a.Encode(source)
	assert.NoError(t, err)
	assert.Equal(t, source, target)
}

func TestDecodeAssembler_Add(t *testing.T) {
	a := NewDecodeAssembler[any, any]()
	a.Add(DecodeCompilerFunc[any](func(typ reflect.Type) (Decoder[any, unsafe.Pointer], error) {
		return nil, nil
	}))

	assert.Equal(t, 1, a.Len())
}

func TestDecodeAssembler_Compile(t *testing.T) {
	a := NewDecodeAssembler[any, any]()
	a.Add(DecodeCompilerFunc[any](func(typ reflect.Type) (Decoder[any, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.String {
			return DecodeFunc[any, unsafe.Pointer](func(source any, target unsafe.Pointer) error {
				return nil
			}), nil
		}
		return nil, errors.WithStack(ErrUnsupportedType)
	}))

	source := "test"
	d, err := a.Compile(reflect.TypeOf(&source))
	assert.NoError(t, err)
	assert.NotNil(t, d)
}

func TestDecodeAssembler_Decode(t *testing.T) {
	a := NewDecodeAssembler[any, any]()
	a.Add(DecodeCompilerFunc[any](func(typ reflect.Type) (Decoder[any, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.String {
			return DecodeFunc[any, unsafe.Pointer](func(source any, target unsafe.Pointer) error {
				if s, ok := source.(*string); ok {
					*(*string)(target) = *s
					return nil
				}
				return errors.WithStack(ErrUnsupportedType)
			}), nil
		}
		return nil, errors.WithStack(ErrUnsupportedType)
	}))

	source := faker.UUIDHyphenated()
	target := ""

	err := a.Decode(&source, &target)
	assert.NoError(t, err)
	assert.Equal(t, source, target)
}
