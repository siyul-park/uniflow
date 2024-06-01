package object

import (
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Error represents an error object.
type Error struct {
	value error
}

var _ Object = (*Error)(nil)
var _ error = (*Error)(nil)

// NewError creates a new Error instance.
func NewError(value error) *Error {
	return &Error{value: value}
}

// String returns the error message as a string.
func (e *Error) Error() string {
	return e.value.Error()
}

// Kind returns the kind of the value.
func (e *Error) Kind() Kind {
	return KindError
}

// Hash calculates and returns the hash code using FNV-1a algorithm.
func (e *Error) Hash() uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(e.value.Error()))
	return h.Sum64()
}

// Interface converts Error to its underlying error.
func (e *Error) Interface() any {
	return e.value
}

// Equal checks if two Error instances are equal.
func (e *Error) Equal(other Object) bool {
	if o, ok := other.(*Error); ok {
		return e.value.Error() == o.value.Error()
	}
	return false
}

// Compare checks whether another Object is equal to this Error instance.
func (e *Error) Compare(other Object) int {
	if o, ok := other.(*Error); ok {
		return compare(e.Error(), o.Error())
	}
	return compare(e.Kind(), KindOf(other))
}

func newErrorEncoder() encoding.EncodeCompiler[Object] {
	return encoding.EncodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[unsafe.Pointer, Object], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().ConvertibleTo(reflect.TypeOf((*error)(nil)).Elem()) {
			return encoding.EncodeFunc[unsafe.Pointer, Object](func(source unsafe.Pointer) (Object, error) {
				t := reflect.NewAt(typ.Elem(), source).Elem().Interface().(error)
				return NewError(t), nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newErrorDecoder() encoding.DecodeCompiler[Object] {
	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().ConvertibleTo(reflect.TypeOf((*error)(nil)).Elem()) {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(*Error); ok {
						t := reflect.NewAt(typ.Elem(), target)
						t.Elem().Set(reflect.ValueOf(s.Interface()))
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(*Error); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
