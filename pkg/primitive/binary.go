package primitive

import (
	"bytes"
	"encoding"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	encoding2 "github.com/siyul-park/uniflow/pkg/encoding"
)

// Binary is a representation of a []byte.
type Binary []byte

var _ Value = (Binary)(nil)

// NewBinary creates a new Binary instance.
func NewBinary(value []byte) Binary {
	return value
}

// Len returns the length of the binary data.
func (b Binary) Len() int {
	return len(b)
}

// Get returns the byte at the specified index.
func (b Binary) Get(index int) byte {
	if index >= len(b) {
		return 0
	}
	return b[index]
}

// Bytes returns the raw byte slice.
func (b Binary) Bytes() []byte {
	return b
}

// Kind returns the type of the binary data.
func (b Binary) Kind() Kind {
	return KindBinary
}

// Compare compares two Binary values.
func (b Binary) Compare(v Value) int {
	if other, ok := v.(Binary); ok {
		return bytes.Compare(b.Bytes(), other.Bytes())
	}
	if b.Kind() > v.Kind() {
		return 1
	}
	return -1
}

// Interface converts Binary to a byte slice.
func (b Binary) Interface() any {
	return []byte(b)
}

func newBinaryEncoder() encoding2.Encoder[any, Value] {
	return encoding2.EncoderFunc[any, Value](func(source any) (Value, error) {
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

func newBinaryDecoder() encoding2.Decoder[Value, any] {
	return encoding2.DecoderFunc[Value, any](func(source Value, target any) error {
		if s, ok := source.(Binary); ok {
			if t, ok := target.(encoding.BinaryUnmarshaler); ok {
				return t.UnmarshalBinary(s.Bytes())
			} else if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				if (t.Elem().Kind() == reflect.Slice || t.Elem().Kind() == reflect.Array) && t.Elem().Type().Elem().Kind() == reflect.Uint8 {
					for i := 0; i < s.Len(); i++ {
						if t.Elem().Len() < i+1 {
							if t.Elem().Kind() == reflect.Slice {
								t.Elem().Set(reflect.Append(t.Elem(), reflect.ValueOf(s.Get(i))).Convert(t.Elem().Type()))
							} else {
								return errors.WithMessage(encoding2.ErrInvalidValue, fmt.Sprintf("index(%d) is exceeded len(%d)", i, t.Elem().Len()))
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
