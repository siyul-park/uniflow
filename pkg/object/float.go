package object

import (
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"golang.org/x/exp/constraints"
)

// Float is an interface representing a floating-point number.
type Float struct {
	value float64
}

var _ Object = (*Float)(nil)

// NewFloat returns a new Float instance.
func NewFloat(value float64) *Float {
	return &Float{value: value}
}

// Float returns the raw representation of the float.
func (f *Float) Float() float64 {
	return f.value
}

// Kind returns the type of the float data.
func (f *Float) Kind() Kind {
	return KindFloat
}

// Hash calculates and returns the hash code.
func (f *Float) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[8]byte)(unsafe.Pointer(&f.value))[:])
	return h.Sum64()
}

// Interface converts Float to a float64.
func (f *Float) Interface() any {
	return f.value
}

// Equal checks whether two Float instances are equal.
func (f *Float) Equal(other Object) bool {
	if o, ok := other.(*Float); ok {
		return f.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Float instance.
func (f *Float) Compare(other Object) int {
	if o, ok := other.(*Float); ok {
		return compare(f.value, o.value)
	}
	return compare(f.Kind(), KindOf(other))
}

func newFloatEncoder() encoding.EncodeCompiler[Object] {
	return encoding.EncodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[unsafe.Pointer, Object], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Float32 {
				return newFloatEncoderWithType[float32](), nil
			} else if typ.Elem().Kind() == reflect.Float64 {
				return newFloatEncoderWithType[float64](), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newFloatDecoder() encoding.DecodeCompiler[Object] {
	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Float32 {
				return newFloatDecoderWithType[float32](), nil
			} else if typ.Elem().Kind() == reflect.Float64 {
				return newFloatDecoderWithType[float64](), nil
			} else if typ.Elem().Kind() == reflect.Int {
				return newFloatDecoderWithType[int](), nil
			} else if typ.Elem().Kind() == reflect.Int8 {
				return newFloatDecoderWithType[int8](), nil
			} else if typ.Elem().Kind() == reflect.Int16 {
				return newFloatDecoderWithType[int16](), nil
			} else if typ.Elem().Kind() == reflect.Int32 {
				return newFloatDecoderWithType[int32](), nil
			} else if typ.Elem().Kind() == reflect.Int64 {
				return newFloatDecoderWithType[int64](), nil
			} else if typ.Elem().Kind() == reflect.Uint {
				return newFloatDecoderWithType[uint](), nil
			} else if typ.Elem().Kind() == reflect.Uint8 {
				return newFloatDecoderWithType[uint8](), nil
			} else if typ.Elem().Kind() == reflect.Uint16 {
				return newFloatDecoderWithType[uint16](), nil
			} else if typ.Elem().Kind() == reflect.Uint32 {
				return newFloatDecoderWithType[uint32](), nil
			} else if typ.Elem().Kind() == reflect.Uint64 {
				return newFloatDecoderWithType[uint64](), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(*Float); ok {
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

func newFloatEncoderWithType[T constraints.Integer | constraints.Float]() encoding.Encoder[unsafe.Pointer, Object] {
	return encoding.EncodeFunc[unsafe.Pointer, Object](func(source unsafe.Pointer) (Object, error) {
		t := *(*T)(source)
		return NewFloat(float64(t)), nil
	})
}

func newFloatDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Decoder[Object, unsafe.Pointer] {
	return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
		if s, ok := source.(*Float); ok {
			*(*T)(target) = T(s.Float())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
