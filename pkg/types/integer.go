package types

import (
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"golang.org/x/exp/constraints"
)

// Integer is an interface representing an integer.
type Integer interface {
	Object
	Int() int64
}

// Int represents an int type.
type Int struct {
	value int
}

// Int8 represents an int8 type.
type Int8 struct {
	value int8
}

// Int16 represents an int16 type.
type Int16 struct {
	value int16
}

// Int32 represents an int32 type.
type Int32 struct {
	value int32
}

// Int64 represents an int64 type.
type Int64 struct {
	value int64
}

var _ Integer = Int{}
var _ Integer = Int8{}
var _ Integer = Int16{}
var _ Integer = Int32{}
var _ Integer = Int64{}

// NewInt returns a new Int instance.
func NewInt(value int) Int {
	return Int{value: value}
}

// Int returns the raw representation of the integer.
func (i Int) Int() int64 {
	return int64(i.value)
}

// Kind returns the type of the integer data.
func (i Int) Kind() Kind {
	return KindInt
}

// Hash calculates and returns the hash code.
func (i Int) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[unsafe.Sizeof(i.value)]byte)(unsafe.Pointer(&i.value))[:])
	return h.Sum64()
}

// Interface converts Int to an int.
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

// NewInt8 returns a new Int8 instance.
func NewInt8(value int8) Int8 {
	return Int8{value: value}
}

// Int returns the raw representation of the integer.
func (i Int8) Int() int64 {
	return int64(i.value)
}

// Kind returns the type of the integer data.
func (i Int8) Kind() Kind {
	return KindInt8
}

// Hash calculates and returns the hash code.
func (i Int8) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[1]byte)(unsafe.Pointer(&i.value))[:])
	return h.Sum64()
}

// Interface converts Int8 to an int8.
func (i Int8) Interface() any {
	return i.value
}

// Equal checks whether two Int8 instances are equal.
func (i Int8) Equal(other Object) bool {
	if o, ok := other.(Int8); ok {
		return i.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Int8 instance.
func (i Int8) Compare(other Object) int {
	if o, ok := other.(Int8); ok {
		return compare(i.value, o.value)
	}
	return compare(i.Kind(), KindOf(other))
}

// NewInt16 returns a new Int16 instance.
func NewInt16(value int16) Int16 {
	return Int16{value: value}
}

// Int returns the raw representation of the integer.
func (i Int16) Int() int64 {
	return int64(i.value)
}

// Kind returns the type of the integer data.
func (i Int16) Kind() Kind {
	return KindInt16
}

// Hash calculates and returns the hash code.
func (i Int16) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[2]byte)(unsafe.Pointer(&i.value))[:])
	return h.Sum64()
}

// Interface converts Int16 to an int16.
func (i Int16) Interface() any {
	return i.value
}

// Equal checks whether two Int16 instances are equal.
func (i Int16) Equal(other Object) bool {
	if o, ok := other.(Int16); ok {
		return i.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Int16 instance.
func (i Int16) Compare(other Object) int {
	if o, ok := other.(Int16); ok {
		return compare(i.value, o.value)
	}
	return compare(i.Kind(), KindOf(other))
}

// NewInt32 returns a new Int32 instance.
func NewInt32(value int32) Int32 {
	return Int32{value: value}
}

// Int returns the raw representation of the integer.
func (i Int32) Int() int64 {
	return int64(i.value)
}

// Kind returns the type of the integer data.
func (i Int32) Kind() Kind {
	return KindInt32
}

// Hash calculates and returns the hash code.
func (i Int32) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[4]byte)(unsafe.Pointer(&i.value))[:])
	return h.Sum64()
}

// Interface converts Int32 to an int32.
func (i Int32) Interface() any {
	return i.value
}

// Equal checks whether two Int32 instances are equal.
func (i Int32) Equal(other Object) bool {
	if o, ok := other.(Int32); ok {
		return i.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Int32 instance.
func (i Int32) Compare(other Object) int {
	if o, ok := other.(Int32); ok {
		return compare(i.value, o.value)
	}
	return compare(i.Kind(), KindOf(other))
}

// NewInt64 returns a new Int64 instance.
func NewInt64(value int64) Int64 {
	return Int64{value: value}
}

// Int returns the raw representation of the integer.
func (i Int64) Int() int64 {
	return i.value
}

// Kind returns the type of the integer data.
func (i Int64) Kind() Kind {
	return KindInt64
}

// Hash calculates and returns the hash code.
func (i Int64) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[8]byte)(unsafe.Pointer(&i.value))[:])
	return h.Sum64()
}

// Interface converts Int64 to an int64.
func (i Int64) Interface() any {
	return i.value
}

// Equal checks whether two Int64 instances are equal.
func (i Int64) Equal(other Object) bool {
	if o, ok := other.(Int64); ok {
		return i.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Int64 instance.
func (i Int64) Compare(other Object) int {
	if o, ok := other.(Int64); ok {
		return compare(i.value, o.value)
	}
	return compare(i.Kind(), KindOf(other))
}

func NewIntegerEncoder() encoding.EncodeCompiler[any, Object] {
	return encoding.EncodeCompilerFunc[any, Object](func(typ reflect.Type) (encoding.Encoder[any, Object], error) {
		if typ.Kind() == reflect.Int {
			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				if s, ok := source.(int); ok {
					return NewInt(s), nil
				} else {
					return NewInt(int(reflect.ValueOf(source).Int())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Int8 {
			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				if s, ok := source.(int8); ok {
					return NewInt8(s), nil
				} else {
					return NewInt8(int8(reflect.ValueOf(source).Int())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Int16 {
			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				if s, ok := source.(int16); ok {
					return NewInt16(s), nil
				} else {
					return NewInt16(int16(reflect.ValueOf(source).Int())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Int32 {
			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				if s, ok := source.(int32); ok {
					return NewInt32(s), nil
				} else {
					return NewInt32(int32(reflect.ValueOf(source).Int())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Int64 {
			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				if s, ok := source.(int64); ok {
					return NewInt64(s), nil
				} else {
					return NewInt64(reflect.ValueOf(source).Int()), nil
				}
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}

func NewIntegerDecoder() encoding.DecodeCompiler[Object] {
	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Float32 {
				return NewIntegerDecoderWithType[float32](), nil
			} else if typ.Elem().Kind() == reflect.Float64 {
				return NewIntegerDecoderWithType[float64](), nil
			} else if typ.Elem().Kind() == reflect.Int {
				return NewIntegerDecoderWithType[int](), nil
			} else if typ.Elem().Kind() == reflect.Int8 {
				return NewIntegerDecoderWithType[int8](), nil
			} else if typ.Elem().Kind() == reflect.Int16 {
				return NewIntegerDecoderWithType[int16](), nil
			} else if typ.Elem().Kind() == reflect.Int32 {
				return NewIntegerDecoderWithType[int32](), nil
			} else if typ.Elem().Kind() == reflect.Int64 {
				return NewIntegerDecoderWithType[int64](), nil
			} else if typ.Elem().Kind() == reflect.Uint {
				return NewIntegerDecoderWithType[uint](), nil
			} else if typ.Elem().Kind() == reflect.Uint8 {
				return NewIntegerDecoderWithType[uint8](), nil
			} else if typ.Elem().Kind() == reflect.Uint16 {
				return NewIntegerDecoderWithType[uint16](), nil
			} else if typ.Elem().Kind() == reflect.Uint32 {
				return NewIntegerDecoderWithType[uint32](), nil
			} else if typ.Elem().Kind() == reflect.Uint64 {
				return NewIntegerDecoderWithType[uint64](), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(Integer); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding.ErrInvalidArgument)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}

func NewIntegerDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Decoder[Object, unsafe.Pointer] {
	return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
		if s, ok := source.(Integer); ok {
			*(*T)(target) = T(s.Int())
			return nil
		}
		return errors.WithStack(encoding.ErrInvalidArgument)
	})
}
