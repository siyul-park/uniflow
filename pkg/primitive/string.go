package primitive

import (
	"encoding"
	"reflect"

	"github.com/pkg/errors"
	encoding2 "github.com/siyul-park/uniflow/pkg/encoding"
)

type (
	// String is a representation of a string.
	String string
)

var _ Value = (String)("")

// NewString returns a new String.
func NewString(value string) String {
	return String(value)
}

func (o String) Len() int {
	return len([]rune(o))
}

func (o String) Get(index int) rune {
	if index >= len([]rune(o)) {
		return rune(0)
	}
	return []rune(o)[index]
}

// String returns a raw representation.
func (o String) String() string {
	return string(o)
}

func (o String) Kind() Kind {
	return KindString
}

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

func (o String) Interface() any {
	return string(o)
}

// NewStringEncoder is encode string to String.
func NewStringEncoder() encoding2.Encoder[any, Value] {
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

// NewStringDecoder is decode String to string.
func NewStringDecoder() encoding2.Decoder[Value, any] {
	return encoding2.DecoderFunc[Value, any](func(source Value, target any) error {
		if s, ok := source.(String); ok {
			if t, ok := target.(encoding.TextUnmarshaler); ok {
				return t.UnmarshalText([]byte(s.String()))
			} else if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				if t.Elem().Kind() == reflect.String {
					t.Elem().Set(reflect.ValueOf(s.String()))
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
