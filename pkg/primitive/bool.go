package primitive

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Bool is a representation of a bool.
type Bool bool

var _ Value = (Bool)(false)

var (
	TRUE  = NewBool(true)
	FALSE = NewBool(false)
)

// NewBool returns a new Bool.
func NewBool(value bool) Bool {
	return Bool(value)
}

// Bool returns a raw representation.
func (b Bool) Bool() bool {
	return bool(b)
}

// Kind returns the type of the bool data.
func (b Bool) Kind() Kind {
	return KindBool
}

// Compare compares two Bool values.
func (b Bool) Compare(v Value) int {
	if other, ok := v.(Bool); ok {
		switch {
		case b == other:
			return 0
		case b == TRUE:
			return 1
		default:
			return -1
		}
	}
	if b.Kind() > v.Kind() {
		return 1
	}
	return -1
}

// Interface converts Bool to a bool.
func (b Bool) Interface() any {
	return bool(b)
}

func newBoolEncoder() encoding.Encoder[any, Value] {
	return encoding.EncoderFunc[any, Value](func(source any) (Value, error) {
		if s := reflect.ValueOf(source); s.Kind() == reflect.Bool {
			return NewBool(s.Bool()), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newBoolDecoder() encoding.Decoder[Value, any] {
	return encoding.DecoderFunc[Value, any](func(source Value, target any) error {
		if s, ok := source.(Bool); ok {
			if t := reflect.ValueOf(target); t.Kind() == reflect.Ptr {
				switch {
				case t.Elem().Kind() == reflect.Bool:
					t.Elem().Set(reflect.ValueOf(s.Bool()).Convert(t.Elem().Type()))
					return nil
				case t.Elem().Type() == typeAny:
					t.Elem().Set(reflect.ValueOf(s.Interface()))
					return nil
				}
			}
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
