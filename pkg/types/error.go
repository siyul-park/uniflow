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

var _ Object = (Error)(nil)
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
func (e Error) Equal(other Object) bool {
	if o, ok := other.(Error); ok {
		return e.value.Error() == o.value.Error()
	}
	return false
}

// Compare checks whether another Object is equal to this Error instance.
func (e Error) Compare(other Object) int {
	if o, ok := other.(Error); ok {
		return compare(e.Error(), o.Error())
	}
	return compare(e.Kind(), KindOf(other))
}

func newErrorEncoder() encoding.EncodeCompiler[any, Object] {
	typeError := reflect.TypeOf((*error)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, Object](func(typ reflect.Type) (encoding.Encoder[any, Object], error) {
		if typ.ConvertibleTo(typeError) {
			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				s := source.(error)
				return NewError(s), nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}

func newErrorDecoder() encoding.DecodeCompiler[Object] {
	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().ConvertibleTo(reflect.TypeOf((*error)(nil)).Elem()) {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(Error); ok {
						t := reflect.NewAt(typ.Elem(), target)
						t.Elem().Set(reflect.ValueOf(s.Interface()))
						return nil
					}
					return errors.WithStack(encoding.ErrInvalidArgument)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
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
