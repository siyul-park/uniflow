package types

import (
	"encoding"
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	encoding2 "github.com/siyul-park/uniflow/encoding"
)

// Error represents an error types.
type Error = *_error

type _error struct {
	value error
}

var _ Value = (Error)(nil)
var _ error = (Error)(nil)
var _ encoding.TextMarshaler = (Error)(nil)
var _ encoding.TextUnmarshaler = (Error)(nil)

// NewError creates a new Error instance.
func NewError(value error) Error {
	return &_error{value: value}
}

// String returns the error message as a string.
func (e Error) Error() string {
	return e.value.Error()
}

// Unwrap returns its underlying error.
func (e Error) Unwrap() error {
	return e.value
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

// MarshalText implements the encoding.TextMarshaler interface.
func (e Error) MarshalText() ([]byte, error) {
	return []byte(e.Error()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (e Error) UnmarshalText(text []byte) error {
	e.value = errors.New(string(text))
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (e Error) MarshalBinary() (data []byte, err error) {
	return e.MarshalText()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (e Error) UnmarshalBinary(data []byte) error {
	return e.UnmarshalText(data)
}

func newErrorEncoder() encoding2.EncodeCompiler[any, Value] {
	typeError := reflect.TypeOf((*error)(nil)).Elem()

	return encoding2.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding2.Encoder[any, Value], error) {
		if typ != nil && typ.ConvertibleTo(typeError) {
			return encoding2.EncodeFunc(func(source any) (Value, error) {
				s := source.(error)
				return NewError(s), nil
			}), nil
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedType)
	})
}

func newErrorDecoder() encoding2.DecodeCompiler[Value] {
	typeError := reflect.TypeOf((*error)(nil)).Elem()
	typeTextUnmarshaler := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeBinaryUnmarshaler := reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()

	return encoding2.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding2.Decoder[Value, unsafe.Pointer], error) {
		if typ == nil {
			return nil, errors.WithStack(encoding2.ErrUnsupportedType)
		} else if typ.ConvertibleTo(typeTextUnmarshaler) {
			return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
				if s, ok := source.(Error); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.TextUnmarshaler)
					if err := t.UnmarshalText([]byte(s.Error())); err != nil {
						return errors.Wrap(encoding2.ErrUnsupportedValue, err.Error())
					}
					return nil
				}
				return errors.WithStack(encoding2.ErrUnsupportedType)
			}), nil
		} else if typ.ConvertibleTo(typeBinaryUnmarshaler) {
			return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
				if s, ok := source.(Error); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.BinaryUnmarshaler)
					if err := t.UnmarshalBinary([]byte(s.Error())); err != nil {
						return errors.Wrap(encoding2.ErrUnsupportedValue, err.Error())
					}
					return nil
				}
				return errors.WithStack(encoding2.ErrUnsupportedType)
			}), nil
		} else if typ.Kind() == reflect.Pointer {
			if typ.Elem().ConvertibleTo(typeError) {
				return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Error); ok {
						t := reflect.NewAt(typ.Elem(), target)
						t.Elem().Set(reflect.ValueOf(s.Interface()))
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.String {
				return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Error); ok {
						*(*string)(target) = s.Error()
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem() == types[KindUnknown] {
				return encoding2.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Error); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedType)
	})
}
