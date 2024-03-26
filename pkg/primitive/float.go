package primitive

import (
	"golang.org/x/exp/constraints"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Float is an interface representing a floating-point number.
type Float interface {
	Value
	Float() float64
}

// Float32 is a representation of a float32.
type Float32 float32

// Float64 is a representation of a float64.
type Float64 float64

var _ Float = (Float32)(0)
var _ Float = (Float64)(0)

// NewFloat32 returns a new Float32.
func NewFloat32(value float32) Float32 {
	return Float32(value)
}

// Float returns the raw representation.
func (f Float32) Float() float64 {
	return float64(f)
}

// Kind returns the type of the float32 data.
func (f Float32) Kind() Kind {
	return KindFloat32
}

// Compare compares two Float32 values.
func (f Float32) Compare(v Value) int {
	return compareAsFloat(f, v)
}

// Interface converts Float32 to a float32.
func (f Float32) Interface() any {
	return float32(f)
}

// NewFloat64 returns a new Float64.
func NewFloat64(value float64) Float64 {
	return Float64(value)
}

// Float returns the raw representation.
func (f Float64) Float() float64 {
	return float64(f)
}

// Kind returns the type of the float64 data.
func (f Float64) Kind() Kind {
	return KindFloat64
}

// Compare compares two Float64 values.
func (f Float64) Compare(v Value) int {
	return compareAsFloat(f, v)
}

// Interface converts Float64 to a float64.
func (f Float64) Interface() any {
	return float64(f)
}

func compareAsFloat(f Float, v Value) int {
	if r, ok := v.(Float); ok {
		return compare[float64](f.Float(), r.Float())
	}
	if r, ok := v.(Integer); ok {
		return compare[float64](f.Float(), float64(r.Int()))
	}
	if r, ok := v.(Uinteger); ok {
		return compare[float64](f.Float(), float64(r.Uint()))
	}
	if f.Kind() > v.Kind() {
		return 1
	}
	return -1
}

func newFloatEncoder() encoding.Compiler[*Value] {
	return encoding.CompilerFunc[*Value](func(typ reflect.Type) (encoding.Decoder[*Value, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Float32 {
				return encoding.DecoderFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*float32)(target)
					*source = NewFloat32(t)

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Float64 {
				return encoding.DecoderFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*float64)(target)
					*source = NewFloat64(t)

					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newFloatDecoder() encoding.Compiler[Value] {
	return encoding.CompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
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
				return encoding.DecoderFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
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

func newFloatDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Decoder[Value, unsafe.Pointer] {
	return encoding.DecoderFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
		if s, ok := source.(Float); ok {
			*(*T)(target) = T(s.Float())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
