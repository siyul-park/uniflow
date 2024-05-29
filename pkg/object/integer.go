package object

import (
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"golang.org/x/exp/constraints"
)

// Integer is an interface representing an integer.
type Integer int64

var _ Object = (Integer)(0)

// NewInteger returns a new Int64.
func NewInteger(value int64) Integer {
	return Integer(value)
}

// Int returns the raw representation.
func (i Integer) Int() int64 {
	return int64(i)
}

// Kind returns the type of the int64 data.
func (i Integer) Kind() Kind {
	return KindInteger
}

// Compare compares two Int64 values.
func (i Integer) Compare(v Object) int {
	if r, ok := v.(Integer); ok {
		return compare(i.Int(), r.Int())
	}
	if r, ok := v.(UInteger); ok {
		return compare(i.Int(), int64(r.Uint()))
	}
	if r, ok := v.(Float); ok {
		return compare(float64(i.Int()), r.Float())
	}
	if KindOf(i) > KindOf(v) {
		return 1
	}
	return -1
}

// Hash calculates and returns the hash code.
func (i Integer) Hash() uint64 {
	return *(*uint64)(unsafe.Pointer(&i))
}

// Interface converts Int64 to an int64.
func (i Integer) Interface() any {
	return int64(i)
}

func NewIntegerEncoder() encoding.Compiler[*Object] {
	return encoding.CompilerFunc[*Object](func(typ reflect.Type) (encoding.Encoder[*Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Int {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*int)(target)
					*source = NewInteger(int64(t))

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Int8 {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*int8)(target)
					*source = NewInteger(int64(t))

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Int16 {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*int16)(target)
					*source = NewInteger(int64(t))

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Int32 {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*int32)(target)
					*source = NewInteger(int64(t))

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Int64 {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*int64)(target)
					*source = NewInteger(t)

					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func NewIntegerDecoder() encoding.Compiler[Object] {
	return encoding.CompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Float32 {
				return NewIntegerDecoderWithType[float32](), nil
			} else if typ.Elem().Kind() == reflect.Float64 {
				return NewIntegerDecoderWithType[float64](), nil
			} else if typ.Elem().Kind() == reflect.Int {
				return NewIntegerDecoderWithType[int](), nil
			} else if typ.Elem().Kind() == reflect.Int8 {
				return NewIntegerDecoderWithType[int8](), nil
			} else if typ.Elem().Kind() == reflect.Int16 {
				return NewIntegerDecoderWithType[int16](), nil
			} else if typ.Elem().Kind() == reflect.Int32 {
				return NewIntegerDecoderWithType[int32](), nil
			} else if typ.Elem().Kind() == reflect.Int64 {
				return NewIntegerDecoderWithType[int64](), nil
			} else if typ.Elem().Kind() == reflect.Uint {
				return NewIntegerDecoderWithType[uint](), nil
			} else if typ.Elem().Kind() == reflect.Uint8 {
				return NewIntegerDecoderWithType[uint8](), nil
			} else if typ.Elem().Kind() == reflect.Uint16 {
				return NewIntegerDecoderWithType[uint16](), nil
			} else if typ.Elem().Kind() == reflect.Uint32 {
				return NewIntegerDecoderWithType[uint32](), nil
			} else if typ.Elem().Kind() == reflect.Uint64 {
				return NewIntegerDecoderWithType[uint64](), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(Integer); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func NewIntegerDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Encoder[Object, unsafe.Pointer] {
	return encoding.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
		if s, ok := source.(Integer); ok {
			*(*T)(target) = T(s.Int())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
