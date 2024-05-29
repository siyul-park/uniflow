package object

import (
	"encoding"
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	encoding2 "github.com/siyul-park/uniflow/pkg/encoding"
)

// String is a representation of a string.
type String string

var _ Object = (String)("")

// NewString returns a new String.
func NewString(value string) String {
	return String(value)
}

// Len returns the length of the string.
func (s String) Len() int {
	return len([]rune(s))
}

// Get returns the rune at the specified index in the string.
func (s String) Get(index int) rune {
	runes := []rune(s)
	if index >= len(runes) {
		return rune(0)
	}
	return runes[index]
}

// String returns the raw string representation.
func (s String) String() string {
	return string(s)
}

// Kind returns the kind of the value.
func (s String) Kind() Kind {
	return KindString
}

// Hash calculates and returns the hash code.
func (s String) Hash() uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

// Compare compares two String values.
func (s String) Compare(v Object) int {
	if r, ok := v.(String); !ok {
		if KindOf(s) > KindOf(v) {
			return 1
		} else {
			return -1
		}
	} else {
		return compare[string](s.String(), r.String())
	}
}

// Interface converts String to its underlying string.
func (o String) Interface() any {
	return string(o)
}

func newStringEncoder() encoding2.Compiler[*Object] {
	typeTextMarshaler := reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

	return encoding2.CompilerFunc[*Object](func(typ reflect.Type) (encoding2.Encoder[*Object, unsafe.Pointer], error) {
		if typ.ConvertibleTo(typeTextMarshaler) {
			return encoding2.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.TextMarshaler)
				if s, err := t.MarshalText(); err != nil {
					return err
				} else {
					*source = NewString(string(s))
				}
				return nil
			}), nil
		} else if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.String {
				return encoding2.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*string)(target)
					*source = NewString(t)
					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedValue)
	})
}

func newStringDecoder() encoding2.Compiler[Object] {
	typeTextUnmarshaler := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeBinaryUnmarshaler := reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()

	return encoding2.CompilerFunc[Object](func(typ reflect.Type) (encoding2.Encoder[Object, unsafe.Pointer], error) {
		if typ.ConvertibleTo(typeTextUnmarshaler) {
			return encoding2.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
				if s, ok := source.(String); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.TextUnmarshaler)
					return t.UnmarshalText([]byte(s.String()))
				}
				return errors.WithStack(encoding2.ErrUnsupportedValue)
			}), nil
		} else if typ.ConvertibleTo(typeBinaryUnmarshaler) {
			return encoding2.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
				if s, ok := source.(String); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.BinaryUnmarshaler)
					return t.UnmarshalBinary([]byte(s.String()))
				}
				return errors.WithStack(encoding2.ErrUnsupportedValue)
			}), nil
		} else if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.String {
				return encoding2.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						*(*string)(target) = s.String()
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedValue)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding2.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
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
