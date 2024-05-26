package primitive

import (
	"bytes"
	"encoding"
	"reflect"
	"unsafe"

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
	if KindOf(b) > KindOf(v) {
		return 1
	}
	return -1
}

// Interface converts Binary to a byte slice.
func (b Binary) Interface() any {
	return []byte(b)
}

func newBinaryEncoder() encoding2.Compiler[*Value] {
	typeBinaryMarshaler := reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()

	return encoding2.CompilerFunc[*Value](func(typ reflect.Type) (encoding2.Encoder[*Value, unsafe.Pointer], error) {
		if typ.ConvertibleTo(typeBinaryMarshaler) {
			return encoding2.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.BinaryMarshaler)
				if s, err := t.MarshalBinary(); err != nil {
					return err
				} else {
					*source = NewBinary(s)
				}
				return nil
			}), nil
		} else if typ.Kind() == reflect.Pointer {
			if (typ.Elem().Kind() == reflect.Slice || typ.Elem().Kind() == reflect.Array) && typ.Elem().Elem().Kind() == reflect.Uint8 {
				return encoding2.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := reflect.NewAt(typ.Elem(), target).Elem()
					*source = NewBinary(t.Bytes())
					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedValue)
	})
}

func newBinaryDecoder() encoding2.Compiler[Value] {
	typeBinaryUnmarshaler := reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
	typeTextUnmarshaler := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

	return encoding2.CompilerFunc[Value](func(typ reflect.Type) (encoding2.Encoder[Value, unsafe.Pointer], error) {
		if typ.ConvertibleTo(typeBinaryUnmarshaler) {
			return encoding2.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				if s, ok := source.(Binary); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.BinaryUnmarshaler)
					return t.UnmarshalBinary(s.Bytes())
				}
				return errors.WithStack(encoding2.ErrUnsupportedValue)
			}), nil
		} else if typ.ConvertibleTo(typeTextUnmarshaler) {
			return encoding2.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				if s, ok := source.(Binary); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.TextUnmarshaler)
					return t.UnmarshalText(s.Bytes())
				}
				return errors.WithStack(encoding2.ErrUnsupportedValue)
			}), nil
		} else if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Slice && typ.Elem().Elem().Kind() == reflect.Uint8 {
				return encoding2.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Binary); ok {
						t := reflect.NewAt(typ.Elem(), target).Elem()
						t.Set(reflect.AppendSlice(t, reflect.ValueOf(s.Bytes()).Convert(t.Type())))
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedValue)
				}), nil
			} else if typ.Elem().Kind() == reflect.Array && typ.Elem().Elem().Kind() == reflect.Uint8 {
				return encoding2.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Binary); ok {
						t := reflect.NewAt(typ.Elem(), target).Elem()
						reflect.Copy(t, reflect.ValueOf(s.Bytes()).Convert(t.Type()))
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedValue)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding2.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Binary); ok {
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
