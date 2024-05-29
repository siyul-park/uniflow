package object

import (
	"reflect"
	"unsafe"

	"golang.org/x/exp/constraints"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Float is an interface representing a floating-point number.
type Float float64

var _ Object = (Float)(0)

// NewFloat returns a new Float64.
func NewFloat(value float64) Float {
	return Float(value)
}

// Float returns the raw representation.
func (f Float) Float() float64 {
	return float64(f)
}

// Kind returns the type of the float64 data.
func (f Float) Kind() Kind {
	return KindFloat
}

// Compare compares two Float64 values.
func (f Float) Compare(v Object) int {
	if r, ok := v.(Float); ok {
		return compare(f.Float(), r.Float())
	}
	if r, ok := v.(Integer); ok {
		return compare(f.Float(), float64(r.Int()))
	}
	if r, ok := v.(UInteger); ok {
		return compare(f.Float(), float64(r.Uint()))
	}
	if KindOf(f) > KindOf(v) {
		return 1
	}
	return -1
}

// Hash calculates and returns the hash code.
func (f Float) Hash() uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

// Interface converts Float64 to a float64.
func (f Float) Interface() any {
	return float64(f)
}

func newFloatEncoder() encoding.Compiler[*Object] {
	return encoding.CompilerFunc[*Object](func(typ reflect.Type) (encoding.Encoder[*Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Float32 {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*float32)(target)
					*source = NewFloat(float64(t))

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Float64 {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := *(*float64)(target)
					*source = NewFloat(t)

					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newFloatDecoder() encoding.Compiler[Object] {
	return encoding.CompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[Object, unsafe.Pointer], error) {
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
				return encoding.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
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

func newFloatDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Encoder[Object, unsafe.Pointer] {
	return encoding.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
		if s, ok := source.(Float); ok {
			*(*T)(target) = T(s.Float())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
