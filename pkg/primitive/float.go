package primitive

import (
	"reflect"

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

func newFloatEncoder() encoding.Encoder[any, Value] {
	return encoding.EncoderFunc[any, Value](func(source any) (Value, error) {
		switch s := reflect.ValueOf(source); s.Kind() {
		case reflect.Float32:
			return NewFloat32(float32(s.Float())), nil
		case reflect.Float64:
			return NewFloat64(s.Float()), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newFloatDecoder() encoding.Decoder[Value, any] {
	return encoding.DecoderFunc[Value, any](func(source Value, target any) error {
		if s, ok := source.(Float); ok {
			if t := reflect.ValueOf(target); t.Kind() == reflect.Ptr {
				switch t.Elem().Kind() {
				case reflect.Float32:
					t.Elem().Set(reflect.ValueOf(float32(s.Float())))
				case reflect.Float64:
					t.Elem().Set(reflect.ValueOf(s.Float()))
				case reflect.Int:
					t.Elem().Set(reflect.ValueOf(int(s.Float())))
				case reflect.Int8:
					t.Elem().Set(reflect.ValueOf(int8(s.Float())))
				case reflect.Int16:
					t.Elem().Set(reflect.ValueOf(int16(s.Float())))
				case reflect.Int32:
					t.Elem().Set(reflect.ValueOf(int32(s.Float())))
				case reflect.Int64:
					t.Elem().Set(reflect.ValueOf(int32(s.Float())))
				case reflect.Uint:
					t.Elem().Set(reflect.ValueOf(uint(s.Float())))
				case reflect.Uint8:
					t.Elem().Set(reflect.ValueOf(uint8(s.Float())))
				case reflect.Uint16:
					t.Elem().Set(reflect.ValueOf(uint16(s.Float())))
				case reflect.Uint32:
					t.Elem().Set(reflect.ValueOf(uint32(s.Float())))
				case reflect.Uint64:
					t.Elem().Set(reflect.ValueOf(uint32(s.Float())))
				case reflect.Bool:
					t.Elem().Set(reflect.ValueOf(s.Float() != 0))
				default:
					if t.Type() == typeAny {
						t.Elem().Set(reflect.ValueOf(s.Interface()))
					} else {
						return errors.WithStack(encoding.ErrUnsupportedValue)
					}
				}
				return nil
			}
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
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
