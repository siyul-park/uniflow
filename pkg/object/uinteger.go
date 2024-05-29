package object

import (
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"golang.org/x/exp/constraints"
)

// UInteger is an interface representing an unsigned integer.
type UInteger uint64

var _ Object = (UInteger)(0)

// NewUInteger returns a new Uint64.
func NewUInteger(value uint64) UInteger {
	return UInteger(value)
}

// Uint returns the raw representation.
func (u UInteger) Uint() uint64 {
	return uint64(u)
}

// Kind returns the type of the uint64 data.
func (u UInteger) Kind() Kind {
	return KindUInteger
}

// Compare compares two Uint64 values.
func (u UInteger) Compare(v Object) int {
	if r, ok := v.(UInteger); ok {
		return compare(u.Uint(), r.Uint())
	}
	if r, ok := v.(Integer); ok {
		return compare(int64(u.Uint()), r.Int())
	}
	if r, ok := v.(Float); ok {
		return compare(float64(u.Uint()), r.Float())
	}
	if KindOf(u) > KindOf(v) {
		return 1
	}
	return -1
}

// Hash calculates and returns the hash code.
func (u UInteger) Hash() uint64 {
	return uint64(u)
}

// Interface converts Uint64 to a uint64.
func (u UInteger) Interface() any {
	return uint64(u)
}

func newUIntegerEncoder() encoding.Compiler[*Object] {
	return encoding.CompilerFunc[*Object](func(typ reflect.Type) (encoding.Encoder[*Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Uint {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*uint)(target)
					*source = NewUInteger(uint64(t))

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint8 {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*uint8)(target)
					*source = NewUInteger(uint64(t))

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint16 {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*uint16)(target)
					*source = NewUInteger(uint64(t))

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint32 {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*uint32)(target)
					*source = NewUInteger(uint64(t))

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint64 {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*uint64)(target)
					*source = NewUInteger(t)

					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newUintegerDecoder() encoding.Compiler[Object] {
	return encoding.CompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Float32 {
				return newUintegerDecoderWithType[float32](), nil
			} else if typ.Elem().Kind() == reflect.Float64 {
				return newUintegerDecoderWithType[float64](), nil
			} else if typ.Elem().Kind() == reflect.Int {
				return newUintegerDecoderWithType[int](), nil
			} else if typ.Elem().Kind() == reflect.Int8 {
				return newUintegerDecoderWithType[int8](), nil
			} else if typ.Elem().Kind() == reflect.Int16 {
				return newUintegerDecoderWithType[int16](), nil
			} else if typ.Elem().Kind() == reflect.Int32 {
				return newUintegerDecoderWithType[int32](), nil
			} else if typ.Elem().Kind() == reflect.Int64 {
				return newUintegerDecoderWithType[int64](), nil
			} else if typ.Elem().Kind() == reflect.Uint {
				return newUintegerDecoderWithType[uint](), nil
			} else if typ.Elem().Kind() == reflect.Uint8 {
				return newUintegerDecoderWithType[uint8](), nil
			} else if typ.Elem().Kind() == reflect.Uint16 {
				return newUintegerDecoderWithType[uint16](), nil
			} else if typ.Elem().Kind() == reflect.Uint32 {
				return newUintegerDecoderWithType[uint32](), nil
			} else if typ.Elem().Kind() == reflect.Uint64 {
				return newUintegerDecoderWithType[uint64](), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(UInteger); ok {
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

func newUintegerDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Encoder[Object, unsafe.Pointer] {
	return encoding.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
		if s, ok := source.(UInteger); ok {
			*(*T)(target) = T(s.Uint())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
