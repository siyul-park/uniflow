package primitive

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/internal/encoding"
)

type (
	// Bool is a representation of a bool
	Bool bool
)

var _ Object = (Bool)(false)

var (
	TRUE  = NewBool(true)
	FALSE = NewBool(false)
)

// NewBool returns a new Bool.
func NewBool(value bool) Bool {
	return Bool(value)
}

// Bool returns a raw representation.
func (o Bool) Bool() bool {
	return bool(o)
}

func (o Bool) Kind() Kind {
	return KindBool
}
func (o Bool) Compare(v Object) int {
	if r, ok := v.(Bool); !ok {
		if o.Kind() > v.Kind() {
			return 1
		} else {
			return -1
		}
	} else if o == r {
		return 0
	} else if o == TRUE {
		return 1
	} else {
		return -1
	}
}

func (o Bool) Interface() any {
	return bool(o)
}

// NewBoolEncoder is encode bool to Bool.
func NewBoolEncoder() encoding.Encoder[any, Object] {
	return encoding.EncoderFunc[any, Object](func(source any) (Object, error) {
		if s := reflect.ValueOf(source); s.Kind() == reflect.Bool {
			return NewBool(s.Bool()), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

// NewBoolDecoder is decode Bool to bool.
func NewBoolDecoder() encoding.Decoder[Object, any] {
	return encoding.DecoderFunc[Object, any](func(source Object, target any) error {
		if s, ok := source.(Bool); ok {
			if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				if t.Elem().Kind() == reflect.Bool {
					t.Elem().Set(reflect.ValueOf(s.Bool()))
					return nil
				} else if t.Elem().Type() == typeAny {
					t.Elem().Set(reflect.ValueOf(s.Interface()))
					return nil
				}
			}
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
