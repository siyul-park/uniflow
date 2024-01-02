package primitive

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Uinteger is an interface representing an unsigned integer.
type Uinteger interface {
	Value
	Uint() uint64
}

// Uint is a representation of a uint.
type Uint uint

// Uint8 is a representation of a uint8.
type Uint8 uint8

// Uint16 is a representation of a uint16.
type Uint16 uint16

// Uint32 is a representation of a uint32.
type Uint32 uint32

// Uint64 is a representation of a uint64.
type Uint64 uint64

var _ Uinteger = (Uint)(0)
var _ Uinteger = (Uint8)(0)
var _ Uinteger = (Uint16)(0)
var _ Uinteger = (Uint32)(0)
var _ Uinteger = (Uint64)(0)

// NewUint returns a new Uint.
func NewUint(value uint) Uint {
	return Uint(value)
}

// Uint returns the raw representation.
func (u Uint) Uint() uint64 {
	return uint64(u)
}

// Kind returns the type of the uint data.
func (u Uint) Kind() Kind {
	return KindUint
}

// Compare compares two Uint values.
func (u Uint) Compare(v Value) int {
	return compareAsUinteger(u, v)
}

// Interface converts Uint to a uint.
func (u Uint) Interface() any {
	return uint(u)
}

// NewUint8 returns a new Uint8.
func NewUint8(value uint8) Uint8 {
	return Uint8(value)
}

// Uint returns the raw representation.
func (u Uint8) Uint() uint64 {
	return uint64(u)
}

// Kind returns the type of the uint8 data.
func (u Uint8) Kind() Kind {
	return KindUint8
}

// Compare compares two Uint8 values.
func (u Uint8) Compare(v Value) int {
	return compareAsUinteger(u, v)
}

// Interface converts Uint8 to a uint8.
func (u Uint8) Interface() any {
	return uint8(u)
}

// NewUint16 returns a new Uint16.
func NewUint16(value uint16) Uint16 {
	return Uint16(value)
}

// Uint returns the raw representation.
func (u Uint16) Uint() uint64 {
	return uint64(u)
}

// Kind returns the type of the uint16 data.
func (u Uint16) Kind() Kind {
	return KindUint16
}

// Compare compares two Uint16 values.
func (u Uint16) Compare(v Value) int {
	return compareAsUinteger(u, v)
}

// Interface converts Uint16 to a uint16.
func (u Uint16) Interface() any {
	return uint16(u)
}

// NewUint32 returns a new Uint32.
func NewUint32(value uint32) Uint32 {
	return Uint32(value)
}

// Uint returns the raw representation.
func (u Uint32) Uint() uint64 {
	return uint64(u)
}

// Kind returns the type of the uint32 data.
func (u Uint32) Kind() Kind {
	return KindUint32
}

// Compare compares two Uint32 values.
func (u Uint32) Compare(v Value) int {
	return compareAsUinteger(u, v)
}

// Interface converts Uint32 to a uint32.
func (u Uint32) Interface() any {
	return uint32(u)
}

// NewUint64 returns a new Uint64.
func NewUint64(value uint64) Uint64 {
	return Uint64(value)
}

// Uint returns the raw representation.
func (u Uint64) Uint() uint64 {
	return uint64(u)
}

// Kind returns the type of the uint64 data.
func (u Uint64) Kind() Kind {
	return KindUint64
}

// Compare compares two Uint64 values.
func (u Uint64) Compare(v Value) int {
	return compareAsUinteger(u, v)
}

// Interface converts Uint64 to a uint64.
func (u Uint64) Interface() any {
	return uint64(u)
}

func newUintEncoder() encoding.Encoder[any, Value] {
	return encoding.EncoderFunc[any, Value](func(source any) (Value, error) {
		switch s := reflect.ValueOf(source); s.Kind() {
		case reflect.Uint:
			return NewUint(uint(s.Uint())), nil
		case reflect.Uint8:
			return NewUint8(uint8(s.Uint())), nil
		case reflect.Uint16:
			return NewUint16(uint16(s.Uint())), nil
		case reflect.Uint32:
			return NewUint32(uint32(s.Uint())), nil
		case reflect.Uint64:
			return NewUint64(uint64(s.Uint())), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newUintDecoder() encoding.Decoder[Value, any] {
	return encoding.DecoderFunc[Value, any](func(source Value, target any) error {
		if s, ok := source.(Uinteger); ok {
			if t := reflect.ValueOf(target); t.Kind() == reflect.Ptr {
				switch t.Elem().Kind() {
				case reflect.Float32:
					t.Elem().Set(reflect.ValueOf(float32(s.Uint())))
				case reflect.Float64:
					t.Elem().Set(reflect.ValueOf(s.Uint()))
				case reflect.Int:
					t.Elem().Set(reflect.ValueOf(int(s.Uint())))
				case reflect.Int8:
					t.Elem().Set(reflect.ValueOf(int8(s.Uint())))
				case reflect.Int16:
					t.Elem().Set(reflect.ValueOf(int16(s.Uint())))
				case reflect.Int32:
					t.Elem().Set(reflect.ValueOf(int32(s.Uint())))
				case reflect.Int64:
					t.Elem().Set(reflect.ValueOf(int32(s.Uint())))
				case reflect.Uint:
					t.Elem().Set(reflect.ValueOf(uint(s.Uint())))
				case reflect.Uint8:
					t.Elem().Set(reflect.ValueOf(uint8(s.Uint())))
				case reflect.Uint16:
					t.Elem().Set(reflect.ValueOf(uint16(s.Uint())))
				case reflect.Uint32:
					t.Elem().Set(reflect.ValueOf(uint32(s.Uint())))
				case reflect.Uint64:
					t.Elem().Set(reflect.ValueOf(uint64(s.Uint())))
				case reflect.Bool:
					t.Elem().Set(reflect.ValueOf(s.Uint() != 0))
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

func compareAsUinteger(u Uinteger, v Value) int {
	if r, ok := v.(Uinteger); ok {
		return compare[uint64](u.Uint(), r.Uint())
	}
	if r, ok := v.(Integer); ok {
		return compare[int64](int64(u.Uint()), r.Int())
	}
	if r, ok := v.(Float); ok {
		return compare[float64](float64(u.Uint()), r.Float())
	}
	if u.Kind() > v.Kind() {
		return 1
	}
	return -1
}
