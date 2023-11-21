package primitive

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"reflect"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/internal/encoding"
)

type (
	Float interface {
		Object
		Float() float64
	}
	// Float32 is a representation of a float64.
	Float32 float32
	// Float64 is a representation of a float64.
	Float64 float64
)

var _ Float = (Float32)(0)
var _ Float = (Float64)(0)

// NewFloat64 returns a new Float64.
func NewFloat32(value float32) Float32 {
	return Float32(value)
}

// Float returns a raw representation.
func (o Float32) Float() float64 {
	return float64(o)
}

func (o Float32) Kind() Kind {
	return KindFloat32
}

func (o Float32) Hash() uint32 {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], math.Float32bits(float32(o)))

	h := fnv.New32()
	h.Write([]byte{byte(KindFloat32), 0})
	h.Write(buf[:])

	return h.Sum32()
}

func (o Float32) Interface() any {
	return float32(o)
}

// NewFloat64 returns a new Float64.
func NewFloat64(value float64) Float64 {
	return Float64(value)
}

// Float returns a raw representation.
func (o Float64) Float() float64 {
	return float64(o)
}

func (o Float64) Kind() Kind {
	return KindFloat64
}

func (o Float64) Hash() uint32 {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], math.Float64bits(float64(o)))

	h := fnv.New32()
	h.Write([]byte{byte(KindFloat64), 0})
	h.Write(buf[:])

	return h.Sum32()
}

func (o Float64) Interface() any {
	return float64(o)
}

// NewFloatEncoder is encode float to Float.
func NewFloatEncoder() encoding.Encoder[any, Object] {
	return encoding.EncoderFunc[any, Object](func(source any) (Object, error) {
		if s := reflect.ValueOf(source); s.Kind() == reflect.Float32 {
			return NewFloat32(float32(s.Float())), nil
		} else if s := reflect.ValueOf(source); s.Kind() == reflect.Float64 {
			return NewFloat64(float64(s.Float())), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

// NewFloatDecoder is decode Float to float.
func NewFloatDecoder() encoding.Decoder[Object, any] {
	return encoding.DecoderFunc[Object, any](func(source Object, target any) error {
		if s, ok := source.(Float); ok {
			if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				if t.Elem().Kind() == reflect.Float32 {
					t.Elem().Set(reflect.ValueOf(float32(s.Float())))
					return nil
				} else if t.Elem().Kind() == reflect.Float64 {
					t.Elem().Set(reflect.ValueOf(float64(s.Float())))
					return nil
				} else if t.Elem().Kind() == reflect.Int {
					t.Elem().Set(reflect.ValueOf(int(s.Float())))
					return nil
				} else if t.Elem().Kind() == reflect.Int8 {
					t.Elem().Set(reflect.ValueOf(int8(s.Float())))
					return nil
				} else if t.Elem().Kind() == reflect.Int16 {
					t.Elem().Set(reflect.ValueOf(int16(s.Float())))
					return nil
				} else if t.Elem().Kind() == reflect.Int32 {
					t.Elem().Set(reflect.ValueOf(int32(s.Float())))
					return nil
				} else if t.Elem().Kind() == reflect.Int64 {
					t.Elem().Set(reflect.ValueOf(int64(s.Float())))
					return nil
				} else if t.Elem().Type() == typeAny {
					t.Elem().Set(reflect.ValueOf(s.Interface()))
					return nil
				}
			}
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}