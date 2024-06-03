package object

import (
	"encoding"
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	encoding2 "github.com/siyul-park/uniflow/pkg/encoding"
)

// String represents a string.
type String struct {
	value string
}

var _ Object = (*String)(nil)

// NewString creates a new String instance.
func NewString(value string) *String {
	return &String{value: value}
}

// Len returns the length of the string.
func (s *String) Len() int {
	return len([]rune(s.value))
}

// Get returns the rune at the specified index in the string.
func (s *String) Get(index int) rune {
	runes := []rune(s.value)
	if index >= len(runes) {
		return rune(0)
	}
	return runes[index]
}

// String returns the raw string representation.
func (s *String) String() string {
	return s.value
}

// Kind returns the kind of the value.
func (s *String) Kind() Kind {
	return KindString
}

// Hash calculates and returns the hash code using FNV-1a algorithm.
func (s *String) Hash() uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s.value))
	return h.Sum64()
}

// Interface converts String to its underlying string.
func (s *String) Interface() any {
	return s.value
}

// Equal checks if two String instances are equal.
func (s *String) Equal(other Object) bool {
	if o, ok := other.(*String); ok {
		return s.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this String instance.
func (s *String) Compare(other Object) int {
	if o, ok := other.(*String); ok {
		return compare(s.value, o.value)
	}
	return compare(s.Kind(), KindOf(other))
}

func newStringEncoder() encoding2.EncodeCompiler[any, Object] {
	typeTextMarshaler := reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

	return encoding2.EncodeCompilerFunc[any, Object](func(typ reflect.Type) (encoding2.Encoder[any, Object], error) {
		if typ.ConvertibleTo(typeTextMarshaler) {
			return encoding2.EncodeFunc[any, Object](func(source any) (Object, error) {
				s := source.(encoding.TextMarshaler)
				if s, err := s.MarshalText(); err != nil {
					return nil, err
				} else {
					return NewString(string(s)), nil
				}
			}), nil
		} else if typ.Kind() == reflect.String {
			return encoding2.EncodeFunc[any, Object](func(source any) (Object, error) {
				if s, ok := source.(string); ok {
					return NewString(s), nil
				} else {
					return NewString(reflect.ValueOf(source).String()), nil
				}
			}), nil
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedValue)
	})
}

func newStringDecoder() encoding2.DecodeCompiler[Object] {
	typeTextUnmarshaler := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeBinaryUnmarshaler := reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()

	return encoding2.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding2.Decoder[Object, unsafe.Pointer], error) {
		if typ.ConvertibleTo(typeTextUnmarshaler) {
			return encoding2.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
				if s, ok := source.(*String); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.TextUnmarshaler)
					return t.UnmarshalText([]byte(s.String()))
				}
				return errors.WithStack(encoding2.ErrUnsupportedValue)
			}), nil
		} else if typ.ConvertibleTo(typeBinaryUnmarshaler) {
			return encoding2.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
				if s, ok := source.(*String); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.BinaryUnmarshaler)
					return t.UnmarshalBinary([]byte(s.String()))
				}
				return errors.WithStack(encoding2.ErrUnsupportedValue)
			}), nil
		} else if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.String {
				return encoding2.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(*String); ok {
						*(*string)(target) = s.String()
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedValue)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding2.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(*String); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedValue)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedValue)
	})
}
