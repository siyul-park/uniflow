package primitive

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"golang.org/x/exp/constraints"
	"reflect"
	"unsafe"
)

// Integer is an interface representing an integer.
type Integer interface {
	Value
	Int() int64
}

// Int is a representation of an int.
type Int int

// Int8 is a representation of an int8.
type Int8 int8

// Int16 is a representation of an int16.
type Int16 int16

// Int32 is a representation of an int32.
type Int32 int32

// Int64 is a representation of an int64.
type Int64 int64

var _ Integer = (Int)(0)
var _ Integer = (Int8)(0)
var _ Integer = (Int16)(0)
var _ Integer = (Int32)(0)
var _ Integer = (Int64)(0)

// NewInt returns a new Int.
func NewInt(value int) Int {
	return Int(value)
}

// Int returns the raw representation.
func (i Int) Int() int64 {
	return int64(i)
}

// Kind returns the type of the int data.
func (i Int) Kind() Kind {
	return KindInt
}

// Compare compares two Int values.
func (i Int) Compare(v Value) int {
	return compareAsInteger(i, v)
}

// Interface converts Int to an int.
func (i Int) Interface() any {
	return int(i)
}

// NewInt8 returns a new Int8.
func NewInt8(value int8) Int8 {
	return Int8(value)
}

// Int returns the raw representation.
func (i Int8) Int() int64 {
	return int64(i)
}

// Kind returns the type of the int8 data.
func (i Int8) Kind() Kind {
	return KindInt8
}

// Compare compares two Int8 values.
func (i Int8) Compare(v Value) int {
	return compareAsInteger(i, v)
}

// Interface converts Int8 to an int8.
func (i Int8) Interface() any {
	return int8(i)
}

// NewInt16 returns a new Int16.
func NewInt16(value int16) Int16 {
	return Int16(value)
}

// Int returns the raw representation.
func (i Int16) Int() int64 {
	return int64(i)
}

// Kind returns the type of the int16 data.
func (i Int16) Kind() Kind {
	return KindInt16
}

// Compare compares two Int16 values.
func (i Int16) Compare(v Value) int {
	return compareAsInteger(i, v)
}

// Interface converts Int16 to an int16.
func (i Int16) Interface() any {
	return int16(i)
}

// NewInt32 returns a new Int32.
func NewInt32(value int32) Int32 {
	return Int32(value)
}

// Int returns the raw representation.
func (i Int32) Int() int64 {
	return int64(i)
}

// Kind returns the type of the int32 data.
func (i Int32) Kind() Kind {
	return KindInt32
}

// Compare compares two Int32 values.
func (i Int32) Compare(v Value) int {
	return compareAsInteger(i, v)
}

// Interface converts Int32 to an int32.
func (i Int32) Interface() any {
	return int32(i)
}

// NewInt64 returns a new Int64.
func NewInt64(value int64) Int64 {
	return Int64(value)
}

// Int returns the raw representation.
func (i Int64) Int() int64 {
	return int64(i)
}

// Kind returns the type of the int64 data.
func (i Int64) Kind() Kind {
	return KindInt64
}

// Compare compares two Int64 values.
func (i Int64) Compare(v Value) int {
	return compareAsInteger(i, v)
}

// Interface converts Int64 to an int64.
func (i Int64) Interface() any {
	return int64(i)
}

func compareAsInteger(i Integer, v Value) int {
	if r, ok := v.(Integer); ok {
		return compare[int64](i.Int(), r.Int())
	}
	if r, ok := v.(Uinteger); ok {
		return compare[int64](i.Int(), int64(r.Uint()))
	}
	if r, ok := v.(Float); ok {
		return compare[float64](float64(i.Int()), r.Float())
	}
	if i.Kind() > v.Kind() {
		return 1
	}
	return -1
}

func newIntegerEncoder() encoding.Compiler[*Value] {
	return encoding.CompilerFunc[*Value](func(typ reflect.Type) (encoding.Encoder[*Value, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Int {
				return encoding.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*int)(target)
					*source = NewInt(t)

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Int8 {
				return encoding.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*int8)(target)
					*source = NewInt8(t)

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Int16 {
				return encoding.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*int16)(target)
					*source = NewInt16(t)

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Int32 {
				return encoding.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*int32)(target)
					*source = NewInt32(t)

					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Int64 {
				return encoding.EncodeFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
					t := *(*int64)(target)
					*source = NewInt64(t)

					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newIntegerDecoder() encoding.Compiler[Value] {
	return encoding.CompilerFunc[Value](func(typ reflect.Type) (encoding.Encoder[Value, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Float32 {
				return newIntegerDecoderWithType[float32](), nil
			} else if typ.Elem().Kind() == reflect.Float64 {
				return newIntegerDecoderWithType[float64](), nil
			} else if typ.Elem().Kind() == reflect.Int {
				return newIntegerDecoderWithType[int](), nil
			} else if typ.Elem().Kind() == reflect.Int8 {
				return newIntegerDecoderWithType[int8](), nil
			} else if typ.Elem().Kind() == reflect.Int16 {
				return newIntegerDecoderWithType[int16](), nil
			} else if typ.Elem().Kind() == reflect.Int32 {
				return newIntegerDecoderWithType[int32](), nil
			} else if typ.Elem().Kind() == reflect.Int64 {
				return newIntegerDecoderWithType[int64](), nil
			} else if typ.Elem().Kind() == reflect.Uint {
				return newIntegerDecoderWithType[uint](), nil
			} else if typ.Elem().Kind() == reflect.Uint8 {
				return newIntegerDecoderWithType[uint8](), nil
			} else if typ.Elem().Kind() == reflect.Uint16 {
				return newIntegerDecoderWithType[uint16](), nil
			} else if typ.Elem().Kind() == reflect.Uint32 {
				return newIntegerDecoderWithType[uint32](), nil
			} else if typ.Elem().Kind() == reflect.Uint64 {
				return newIntegerDecoderWithType[uint64](), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Integer); ok {
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

func newIntegerDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Encoder[Value, unsafe.Pointer] {
	return encoding.EncodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
		if s, ok := source.(Integer); ok {
			*(*T)(target) = T(s.Int())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
