package primitive

import (
	"encoding"
	"github.com/pkg/errors"
	encoding2 "github.com/siyul-park/uniflow/pkg/encoding"
	"reflect"
	"unsafe"
)

// String is a representation of a string.
type String string

var _ Value = (String)("")

// NewString returns a new String.
func NewString(value string) String {
	return String(value)
}

// Len returns the length of the string.
func (o String) Len() int {
	return len([]rune(o))
}

// Get returns the rune at the specified index in the string.
func (o String) Get(index int) rune {
	runes := []rune(o)
	if index >= len(runes) {
		return rune(0)
	}
	return runes[index]
}

// String returns the raw string representation.
func (o String) String() string {
	return string(o)
}

// Kind returns the kind of the value.
func (o String) Kind() Kind {
	return KindString
}

// Compare compares two String values.
func (o String) Compare(v Value) int {
	if r, ok := v.(String); !ok {
		if o.Kind() > v.Kind() {
			return 1
		} else {
			return -1
		}
	} else {
		return compare[string](o.String(), r.String())
	}
}

// Interface converts String to its underlying string.
func (o String) Interface() any {
	return string(o)
}

func newStringEncoder() encoding2.Compiler[*Value] {
	typeTextMarshaler := reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

	return encoding2.CompilerFunc[*Value](func(typ reflect.Type) (encoding2.Encoder[*Value, unsafe.Pointer], error) {
		if typ.ConvertibleTo(typeTextMarshaler) {
			return encoding2.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
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
				return encoding2.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*string)(target)
					*source = NewString(t)
					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedValue)
	})
}

func newStringDecoder() encoding2.Compiler[Value] {
	typeTextUnmarshaler := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

	return encoding2.CompilerFunc[Value](func(typ reflect.Type) (encoding2.Encoder[Value, unsafe.Pointer], error) {
		if typ.ConvertibleTo(typeTextUnmarshaler) {
			return encoding2.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				if s, ok := source.(String); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.TextUnmarshaler)
					return t.UnmarshalText([]byte(s.String()))
				}
				return errors.WithStack(encoding2.ErrUnsupportedValue)
			}), nil
		} else if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.String {
				return encoding2.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						*(*string)(target) = s.String()
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedValue)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding2.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
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
