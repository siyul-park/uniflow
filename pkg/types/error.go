package types

import (
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Error represents an error types.

type Error = *_error

type _error struct {
	value error
}

var _ Value = (Error)(nil)
var _ error = (Error)(nil)

// NewError creates a new Error instance.
func NewError(value error) Error {
	return &_error{value: value}
}

// String returns the error message as a string.
func (e Error) Error() string {
	return e.value.Error()
}

// Kind returns the kind of the value.
func (e Error) Kind() Kind {
	return KindError
}

// Hash calculates and returns the hash code using FNV-1a algorithm.
func (e Error) Hash() uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(e.value.Error()))
	return h.Sum64()
}

// Interface converts Error to its underlying error.
func (e Error) Interface() any {
	return e.value
}

// Equal checks if two Error instances are equal.
func (e Error) Equal(other Value) bool {
	if o, ok := other.(Error); ok {
		return e.value.Error() == o.value.Error()
	}
	return false
}

// Compare checks whether another Object is equal to this Error instance.
func (e Error) Compare(other Value) int {
	if o, ok := other.(Error); ok {
		return compare(e.Error(), o.Error())
	}
	return compare(e.Kind(), KindOf(other))
}

func newErrorEncoder() encoding.EncodeCompiler[any, Value] {
	typeError := reflect.TypeOf((*error)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ != nil && typ.ConvertibleTo(typeError) {
			return encoding.EncodeFunc[any, Value](func(source any) (Value, error) {
				s := source.(error)
				return NewError(s), nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}

func newErrorDecoder() encoding.DecodeCompiler[Value] {
	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
			if typ.Elem().ConvertibleTo(reflect.TypeOf((*error)(nil)).Elem()) {
				return encoding.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Error); ok {
						t := reflect.NewAt(typ.Elem(), target)
						t.Elem().Set(reflect.ValueOf(s.Interface()))
						return nil
					}
					return errors.WithStack(encoding.ErrInvalidArgument)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Error); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding.ErrInvalidArgument)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}
