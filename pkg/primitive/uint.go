package primitive

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"golang.org/x/exp/constraints"
	"reflect"
	"unsafe"
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

func newUintegerEncoder() encoding.Compiler[*Value] {
	return encoding.CompilerFunc[*Value](func(typ reflect.Type) (encoding.Encoder[*Value, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Uint {
				return encoding.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*uint)(target)
					*source = NewUint(t)

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint8 {
				return encoding.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*uint8)(target)
					*source = NewUint8(t)

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint16 {
				return encoding.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*uint16)(target)
					*source = NewUint16(t)

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint32 {
				return encoding.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*uint32)(target)
					*source = NewUint32(t)

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint64 {
				return encoding.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*uint64)(target)
					*source = NewUint64(t)

					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newUintegerDecoder() encoding.Compiler[Value] {
	return encoding.CompilerFunc[Value](func(typ reflect.Type) (encoding.Encoder[Value, unsafe.Pointer], error) {
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
				return encoding.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Uinteger); ok {
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

func newUintegerDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Encoder[Value, unsafe.Pointer] {
	return encoding.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
		if s, ok := source.(Uinteger); ok {
			*(*T)(target) = T(s.Uint())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
