package types

import (
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"golang.org/x/exp/constraints"
)

// Float is an interface representing a floating-point number.
type Float interface {
	Object
	Float() float64
}

// Float32 represents a float32 type.
type Float32 struct {
	value float32
}

// Float64 represents a float64 type.
type Float64 struct {
	value float64
}

var _ Float = Float32{}
var _ Float = Float64{}

// NewFloat32 returns a new Float32 instance.
func NewFloat32(value float32) Float32 {
	return Float32{value: value}
}

// Float returns the raw representation of the float.
func (f Float32) Float() float64 {
	return float64(f.value)
}

// Kind returns the type of the float data.
func (f Float32) Kind() Kind {
	return KindFloat32
}

// Hash calculates and returns the hash code.
func (f Float32) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[4]byte)(unsafe.Pointer(&f.value))[:])
	return h.Sum64()
}

// Interface converts Float32 to a float32.
func (f Float32) Interface() any {
	return f.value
}

// Equal checks whether two Float32 instances are equal.
func (f Float32) Equal(other Object) bool {
	if o, ok := other.(Float32); ok {
		return f.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Float32 instance.
func (f Float32) Compare(other Object) int {
	if o, ok := other.(Float32); ok {
		return compare(f.value, o.value)
	}
	return compare(f.Kind(), KindOf(other))
}

// NewFloat64 returns a new Float64 instance.
func NewFloat64(value float64) Float64 {
	return Float64{value: value}
}

// Float returns the raw representation of the float.
func (f Float64) Float() float64 {
	return f.value
}

// Kind returns the type of the float data.
func (f Float64) Kind() Kind {
	return KindFloat64
}

// Hash calculates and returns the hash code.
func (f Float64) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[8]byte)(unsafe.Pointer(&f.value))[:])
	return h.Sum64()
}

// Interface converts Float64 to a float64.
func (f Float64) Interface() any {
	return f.value
}

// Equal checks whether two Float64 instances are equal.
func (f Float64) Equal(other Object) bool {
	if o, ok := other.(Float64); ok {
		return f.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Float64 instance.
func (f Float64) Compare(other Object) int {
	if o, ok := other.(Float64); ok {
		return compare(f.value, o.value)
	}
	return compare(f.Kind(), KindOf(other))
}

func newFloatEncoder() encoding.EncodeCompiler[any, Object] {
	return encoding.EncodeCompilerFunc[any, Object](func(typ reflect.Type) (encoding.Encoder[any, Object], error) {
		if typ.Kind() == reflect.Float32 {
			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				if s, ok := source.(float32); ok {
					return NewFloat32(s), nil
				} else {
					return NewFloat32(float32(reflect.ValueOf(source).Float())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Float64 {
			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				if s, ok := source.(float64); ok {
					return NewFloat64(s), nil
				} else {
					return NewFloat64(reflect.ValueOf(source).Float()), nil
				}
			}), nil
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
					if s, ok := source.(Float); ok {
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

func newFloatDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Decoder[Object, unsafe.Pointer] {
	return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
		if s, ok := source.(Float); ok {
			*(*T)(target) = T(s.Float())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
