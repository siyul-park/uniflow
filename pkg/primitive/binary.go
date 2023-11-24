package primitive

import (
	"encoding"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	encoding2 "github.com/siyul-park/uniflow/pkg/encoding"
)

type (
	// Binary is a representation of a []byte.
	Binary []byte
)

var _ Object = (Binary)(nil)

// NewBinary returns a new Binary.
func NewBinary(value []byte) Binary {
	return Binary(value)
}

func (o Binary) Len() int {
	return len([]byte(o))
}

func (o Binary) Get(index int) byte {
	if index >= len([]byte(o)) {
		return 0
	}
	return o[index]
}

// Bytes returns a raw representation.
func (o Binary) Bytes() []byte {
	return []byte(o)
}

func (o Binary) Kind() Kind {
	return KindBinary
}

func (o Binary) Compare(v Object) int {
	if r, ok := v.(Binary); !ok {
		if o.Kind() > v.Kind() {
			return 1
		} else {
			return -1
		}
	} else {
		for i := 0; i < o.Len(); i++ {
			if r.Len() == i {
				return 1
			}

			v1 := o.Get(i)
			v2 := r.Get(i)

			if v1 > v2 {
				return 1
			} else if v1 < v2 {
				return -1
			}
		}
		return 0
	}
}

func (o Binary) Interface() any {
	return []byte(o)
}

// NewBinaryEncoder is encode byte like to Binary.
func NewBinaryEncoder() encoding2.Encoder[any, Object] {
	return encoding2.EncoderFunc[any, Object](func(source any) (Object, error) {
		if s, ok := source.(encoding.BinaryMarshaler); ok {
			if data, err := s.MarshalBinary(); err != nil {
				return nil, err
			} else {
				return NewBinary(data), nil
			}
		} else if s := reflect.ValueOf(source); (s.Kind() == reflect.Slice || s.Kind() == reflect.Array) && s.Type().Elem().Kind() == reflect.Uint8 {
			return NewBinary(s.Bytes()), nil
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedValue)
	})
}

// NewBinaryDecoder is decode Binary to byte like.
func NewBinaryDecoder() encoding2.Decoder[Object, any] {
	return encoding2.DecoderFunc[Object, any](func(source Object, target any) error {
		if s, ok := source.(Binary); ok {
			if t, ok := target.(encoding.BinaryUnmarshaler); ok {
				return t.UnmarshalBinary(s.Bytes())
			} else if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				if (t.Elem().Kind() == reflect.Slice || t.Elem().Kind() == reflect.Array) && t.Elem().Type().Elem().Kind() == reflect.Uint8 {
					for i := 0; i < s.Len(); i++ {
						if t.Elem().Len() < i+1 {
							if t.Elem().Kind() == reflect.Slice {
								t.Elem().Set(reflect.Append(t.Elem(), reflect.ValueOf(s.Get(i))))
							} else {
								return errors.WithMessage(encoding2.ErrUnsupportedValue, fmt.Sprintf("index(%d) is exceeded len(%d)", i, t.Elem().Len()))
							}
						} else {
							t.Elem().Index(i).Set(reflect.ValueOf(s.Get(i)))
						}
					}
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
