package primitive

import (
	"encoding"
	"reflect"

	"github.com/pkg/errors"
	encoding2 "github.com/siyul-park/uniflow/pkg/encoding"
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

func newStringEncoder() encoding2.Encoder[any, Value] {
	return encoding2.EncoderFunc[any, Value](func(source any) (Value, error) {
		if s, ok := source.(encoding.TextMarshaler); ok {
			if text, err := s.MarshalText(); err != nil {
				return nil, err
			} else {
				return NewString(string(text)), nil
			}
		} else if s := reflect.ValueOf(source); s.Kind() == reflect.String {
			return NewString(s.String()), nil
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedValue)
	})
}

func newStringDecoder() encoding2.Decoder[Value, any] {
	return encoding2.DecoderFunc[Value, any](func(source Value, target any) error {
		if s, ok := source.(String); ok {
			if t, ok := target.(encoding.TextUnmarshaler); ok {
				return t.UnmarshalText([]byte(s.String()))
			} else if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				if t.Elem().Kind() == reflect.String {
					t.Elem().Set(reflect.ValueOf(s.String()).Convert(t.Elem().Type()))
					return nil
				} else if t.Elem().Type() == typeAny {
					t.Elem().Set(reflect.ValueOf(s.Interface()))
					return nil
				}
			}
		}
		return errors.WithStack(encoding2.ErrUnsupportedValue)
	})
}
