package types

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	encoding2 "github.com/siyul-park/uniflow/pkg/encoding"
)

// Binary is a representation of a []byte.
type Binary = *binary_

type binary_ struct {
	value []byte
}

var _ Value = (Binary)(nil)

// NewBinary creates a new Binary instance.
func NewBinary(value []byte) Binary {
	return &binary_{
		value: value,
	}
}

// Len returns the length of the binary data.
func (b Binary) Len() int {
	return len(b.value)
}

// Get returns the byte at the specified index.
// If the index is out of bounds, it returns 0.
func (b Binary) Get(index int) byte {
	if index >= len(b.value) {
		return 0
	}
	return b.value[index]
}

// Bytes returns the raw byte slice.
func (b Binary) Bytes() []byte {
	return b.value
}

// Kind returns the type of the binary data.
func (b Binary) Kind() Kind {
	return KindBinary
}

// Hash returns the precomputed hash code.
func (b Binary) Hash() uint64 {
	h := fnv.New64a()
	h.Write(b.value)
	return h.Sum64()
}

// Interface converts Binary to a byte slice.
func (b Binary) Interface() any {
	return b.value
}

// Equal checks whether another Object is equal to this Binary instance.
func (b Binary) Equal(other Value) bool {
	if o, ok := other.(Binary); ok {
		return bytes.Equal(b.value, o.value)
	}
	return false
}

// Compare checks whether another Object is equal to this Binary instance.
func (b Binary) Compare(other Value) int {
	if o, ok := other.(Binary); ok {
		return bytes.Compare(b.Bytes(), o.Bytes())
	}
	return compare(b.Kind(), KindOf(other))
}

func newBinaryEncoder() encoding2.EncodeCompiler[any, Value] {
	typeBinaryMarshaler := reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()

	return encoding2.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding2.Encoder[any, Value], error) {
		if typ != nil && typ.ConvertibleTo(typeBinaryMarshaler) {
			return encoding2.EncodeFunc[any, Value](func(source any) (Value, error) {
				s := source.(encoding.BinaryMarshaler)
				if t, err := s.MarshalBinary(); err != nil {
					return nil, err
				} else {
					return NewBinary(t), nil
				}
			}), nil
		} else if typ != nil && typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Uint8 {
			return encoding2.EncodeFunc[any, Value](func(source any) (Value, error) {
				s := reflect.ValueOf(source)
				return NewBinary(s.Bytes()), nil
			}), nil
		} else if typ != nil && typ.Kind() == reflect.Array && typ.Elem().Kind() == reflect.Uint8 {
			return encoding2.EncodeFunc[any, Value](func(source any) (Value, error) {
				s := reflect.ValueOf(source)
				t := reflect.MakeSlice(reflect.SliceOf(typ.Elem()), s.Len(), s.Len())

				reflect.Copy(t, s)

				return NewBinary(t.Bytes()), nil
			}), nil
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedType)
	})
}

func newBinaryDecoder() encoding2.DecodeCompiler[Value] {
	typeBinaryUnmarshaler := reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
	typeTextUnmarshaler := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

	return encoding2.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding2.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.ConvertibleTo(typeBinaryUnmarshaler) {
			return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				if s, ok := source.(Binary); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.BinaryUnmarshaler)
					return t.UnmarshalBinary(s.Bytes())
				}
				return errors.WithStack(encoding2.ErrUnsupportedType)
			}), nil
		} else if typ != nil && typ.ConvertibleTo(typeTextUnmarshaler) {
			return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				if s, ok := source.(Binary); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.TextUnmarshaler)
					return t.UnmarshalText(s.Bytes())
				}
				return errors.WithStack(encoding2.ErrUnsupportedType)
			}), nil
		} else if typ != nil && typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Slice && typ.Elem().Elem().Kind() == reflect.Uint8 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Binary); ok {
						t := reflect.NewAt(typ.Elem(), target).Elem()
						t.Set(reflect.AppendSlice(t, reflect.ValueOf(s.Bytes()).Convert(t.Type())))
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Array && typ.Elem().Elem().Kind() == reflect.Uint8 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Binary); ok {
						t := reflect.NewAt(typ.Elem(), target).Elem()
						reflect.Copy(t, reflect.ValueOf(s.Bytes()).Convert(t.Type()))
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.String {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Binary); ok {
						*(*string)(target) = base64.StdEncoding.EncodeToString(s.Bytes())
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Binary); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedType)
	})
}
