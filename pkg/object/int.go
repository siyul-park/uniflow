package object

import (
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"golang.org/x/exp/constraints"
)

// Int is an interface representing an integer.
type Int struct {
	value int64
}

var _ Object = Int{}

// NewInt returns a new Int instance.
func NewInt(value int64) Int {
	return Int{value: value}
}

// Int returns the raw representation of the integer.
func (i Int) Int() int64 {
	return i.value
}

// Kind returns the type of the integer data.
func (i Int) Kind() Kind {
	return KindInt
}

// Hash calculates and returns the hash code.
func (i Int) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[8]byte)(unsafe.Pointer(&i.value))[:])
	return h.Sum64()
}

// Interface converts Int to an int64.
func (i Int) Interface() any {
	return i.value
}

// Equal checks whether two Int instances are equal.
func (i Int) Equal(other Object) bool {
	if o, ok := other.(Int); ok {
		return i.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Int instance.
func (i Int) Compare(other Object) int {
	if o, ok := other.(Int); ok {
		return compare(i.value, o.value)
	}
	return compare(i.Kind(), KindOf(other))
}

func NewIntEncoder() encoding.EncodeCompiler[any, Object] {
	return encoding.EncodeCompilerFunc[any, Object](func(typ reflect.Type) (encoding.Encoder[any, Object], error) {
		if typ.Kind() == reflect.Int {
			return newIntEncoderWithType[int](), nil
		} else if typ.Kind() == reflect.Int8 {
			return newIntEncoderWithType[int8](), nil
		} else if typ.Kind() == reflect.Int16 {
			return newIntEncoderWithType[int16](), nil
		} else if typ.Kind() == reflect.Int32 {
			return newIntEncoderWithType[int32](), nil
		} else if typ.Kind() == reflect.Int64 {
			return newIntEncoderWithType[int64](), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func NewIntDecoder() encoding.DecodeCompiler[Object] {
	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Float32 {
				return NewIntDecoderWithType[float32](), nil
			} else if typ.Elem().Kind() == reflect.Float64 {
				return NewIntDecoderWithType[float64](), nil
			} else if typ.Elem().Kind() == reflect.Int {
				return NewIntDecoderWithType[int](), nil
			} else if typ.Elem().Kind() == reflect.Int8 {
				return NewIntDecoderWithType[int8](), nil
			} else if typ.Elem().Kind() == reflect.Int16 {
				return NewIntDecoderWithType[int16](), nil
			} else if typ.Elem().Kind() == reflect.Int32 {
				return NewIntDecoderWithType[int32](), nil
			} else if typ.Elem().Kind() == reflect.Int64 {
				return NewIntDecoderWithType[int64](), nil
			} else if typ.Elem().Kind() == reflect.Uint {
				return NewIntDecoderWithType[uint](), nil
			} else if typ.Elem().Kind() == reflect.Uint8 {
				return NewIntDecoderWithType[uint8](), nil
			} else if typ.Elem().Kind() == reflect.Uint16 {
				return NewIntDecoderWithType[uint16](), nil
			} else if typ.Elem().Kind() == reflect.Uint32 {
				return NewIntDecoderWithType[uint32](), nil
			} else if typ.Elem().Kind() == reflect.Uint64 {
				return NewIntDecoderWithType[uint64](), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(Int); ok {
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

func newIntEncoderWithType[T constraints.Integer | constraints.Float]() encoding.Encoder[any, Object] {
	return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
		if s, ok := source.(T); ok {
			return NewInt(int64(s)), nil
		} else {
			return NewInt(reflect.ValueOf(source).Int()), nil
		}
	})
}

func NewIntDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Decoder[Object, unsafe.Pointer] {
	return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
		if s, ok := source.(Int); ok {
			*(*T)(target) = T(s.Int())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
