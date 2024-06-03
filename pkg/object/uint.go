package object

import (
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"golang.org/x/exp/constraints"
)

// Uint is an interface representing an unsigned integer.
type Uint struct {
	value uint64
}

var _ Object = (*Uint)(nil)

// NewUint returns a new Uint64.
func NewUint(value uint64) *Uint {
	return &Uint{value: value}
}

// Uint returns the raw representation.
func (u *Uint) Uint() uint64 {
	return u.value
}

// Kind returns the type of the uint64 data.
func (u *Uint) Kind() Kind {
	return KindUint
}

// Hash returns the hash code for the uint64 value.
func (u *Uint) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[8]byte)(unsafe.Pointer(&u.value))[:])
	return h.Sum64()
}

// Interface converts Uint64 to a uint64.
func (u *Uint) Interface() any {
	return u.value
}

// Equal checks if two Uint objects are equal.
func (u *Uint) Equal(other Object) bool {
	if o, ok := other.(*Uint); ok {
		return u.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Uint instance.
func (u *Uint) Compare(other Object) int {
	if o, ok := other.(*Uint); ok {
		return compare(u.value, o.value)
	}
	return compare(u.Kind(), KindOf(other))
}

func newUintEncoder() encoding.EncodeCompiler[any, Object] {
	return encoding.EncodeCompilerFunc[any, Object](func(typ reflect.Type) (encoding.Encoder[any, Object], error) {
		if typ.Kind() == reflect.Uint {
			return newUintEncoderWithType[uint](), nil
		} else if typ.Kind() == reflect.Uint8 {
			return newUintEncoderWithType[uint8](), nil
		} else if typ.Kind() == reflect.Uint16 {
			return newUintEncoderWithType[uint16](), nil
		} else if typ.Kind() == reflect.Uint32 {
			return newUintEncoderWithType[uint32](), nil
		} else if typ.Kind() == reflect.Uint64 {
			return newUintEncoderWithType[uint64](), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newUintDecoder() encoding.DecodeCompiler[Object] {
	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Float32 {
				return newUintDecoderWithType[float32](), nil
			} else if typ.Elem().Kind() == reflect.Float64 {
				return newUintDecoderWithType[float64](), nil
			} else if typ.Elem().Kind() == reflect.Int {
				return newUintDecoderWithType[int](), nil
			} else if typ.Elem().Kind() == reflect.Int8 {
				return newUintDecoderWithType[int8](), nil
			} else if typ.Elem().Kind() == reflect.Int16 {
				return newUintDecoderWithType[int16](), nil
			} else if typ.Elem().Kind() == reflect.Int32 {
				return newUintDecoderWithType[int32](), nil
			} else if typ.Elem().Kind() == reflect.Int64 {
				return newUintDecoderWithType[int64](), nil
			} else if typ.Elem().Kind() == reflect.Uint {
				return newUintDecoderWithType[uint](), nil
			} else if typ.Elem().Kind() == reflect.Uint8 {
				return newUintDecoderWithType[uint8](), nil
			} else if typ.Elem().Kind() == reflect.Uint16 {
				return newUintDecoderWithType[uint16](), nil
			} else if typ.Elem().Kind() == reflect.Uint32 {
				return newUintDecoderWithType[uint32](), nil
			} else if typ.Elem().Kind() == reflect.Uint64 {
				return newUintDecoderWithType[uint64](), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(*Uint); ok {
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

func newUintEncoderWithType[T constraints.Integer | constraints.Float]() encoding.Encoder[any, Object] {
	return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
		if s, ok := source.(T); ok {
			return NewUint(uint64(s)), nil
		} else {
			return NewUint(reflect.ValueOf(source).Uint()), nil
		}
	})
}

func newUintDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Decoder[Object, unsafe.Pointer] {
	return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
		if s, ok := source.(*Uint); ok {
			*(*T)(target) = T(s.Uint())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
