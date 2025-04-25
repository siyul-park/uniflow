package types

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"

	"github.com/siyul-park/uniflow/internal/encoding"
)

// Integer is an interface representing an integer.
type Integer interface {
	Value
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

var (
	_ Integer = Int{}
	_ Integer = Int8{}
	_ Integer = Int16{}
	_ Integer = Int32{}
	_ Integer = Int64{}

	_ json.Marshaler = Int{}
	_ json.Marshaler = Int8{}
	_ json.Marshaler = Int16{}
	_ json.Marshaler = Int32{}
	_ json.Marshaler = Int64{}

	_ json.Unmarshaler = (*Int)(nil)
	_ json.Unmarshaler = (*Int8)(nil)
	_ json.Unmarshaler = (*Int16)(nil)
	_ json.Unmarshaler = (*Int32)(nil)
	_ json.Unmarshaler = (*Int64)(nil)
)

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
func (i Int) Equal(other Value) bool {
	if o, ok := other.(Int); ok {
		return i.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Int instance.
func (i Int) Compare(other Value) int {
	if o, ok := other.(Int); ok {
		return compare(i.value, o.value)
	}
	return compare(i.Kind(), KindOf(other))
}

// MarshalJSON implements the encoding.MarshalJSON interface.
func (i Int) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// UnmarshalJSON implements the encoding.UnmarshalJSON interface.
func (i *Int) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &i.value)
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
func (i Int8) Equal(other Value) bool {
	if o, ok := other.(Int8); ok {
		return i.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Int8 instance.
func (i Int8) Compare(other Value) int {
	if o, ok := other.(Int8); ok {
		return compare(i.value, o.value)
	}
	return compare(i.Kind(), KindOf(other))
}

// MarshalJSON implements the encoding.MarshalJSON interface.
func (i Int8) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// UnmarshalJSON implements the encoding.UnmarshalJSON interface.
func (i *Int8) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &i.value)
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
func (i Int16) Equal(other Value) bool {
	if o, ok := other.(Int16); ok {
		return i.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Int16 instance.
func (i Int16) Compare(other Value) int {
	if o, ok := other.(Int16); ok {
		return compare(i.value, o.value)
	}
	return compare(i.Kind(), KindOf(other))
}

// MarshalJSON implements the encoding.MarshalJSON interface.
func (i Int16) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// UnmarshalJSON implements the encoding.UnmarshalJSON interface.
func (i *Int16) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &i.value)
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
func (i Int32) Equal(other Value) bool {
	if o, ok := other.(Int32); ok {
		return i.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Int32 instance.
func (i Int32) Compare(other Value) int {
	if o, ok := other.(Int32); ok {
		return compare(i.value, o.value)
	}
	return compare(i.Kind(), KindOf(other))
}

// MarshalJSON implements the encoding.MarshalJSON interface.
func (i Int32) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// UnmarshalJSON implements the encoding.UnmarshalJSON interface.
func (i *Int32) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &i.value)
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
func (i Int64) Equal(other Value) bool {
	if o, ok := other.(Int64); ok {
		return i.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Int64 instance.
func (i Int64) Compare(other Value) int {
	if o, ok := other.(Int64); ok {
		return compare(i.value, o.value)
	}
	return compare(i.Kind(), KindOf(other))
}

// MarshalJSON implements the encoding.MarshalJSON interface.
func (i Int64) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// UnmarshalJSON implements the encoding.UnmarshalJSON interface.
func (i *Int64) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &i.value)
}

func newIntegerEncoder() encoding.EncodeCompiler[any, Value] {
	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ == nil {
			return nil, errors.WithStack(encoding.ErrUnsupportedType)
		} else if typ.Kind() == reflect.Int {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				if s, ok := source.(int); ok {
					return NewInt(s), nil
				} else {
					return NewInt(int(reflect.ValueOf(source).Int())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Int8 {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				if s, ok := source.(int8); ok {
					return NewInt8(s), nil
				} else {
					return NewInt8(int8(reflect.ValueOf(source).Int())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Int16 {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				if s, ok := source.(int16); ok {
					return NewInt16(s), nil
				} else {
					return NewInt16(int16(reflect.ValueOf(source).Int())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Int32 {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				if s, ok := source.(int32); ok {
					return NewInt32(s), nil
				} else {
					return NewInt32(int32(reflect.ValueOf(source).Int())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Int64 {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				if s, ok := source.(int64); ok {
					return NewInt64(s), nil
				} else {
					return NewInt64(reflect.ValueOf(source).Int()), nil
				}
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newIntegerDecoder() encoding.DecodeCompiler[Value] {
	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
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
			} else if typ.Elem().Kind() == reflect.String {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Integer); ok {
						*(*string)(target) = fmt.Sprint(s.Interface())
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedType)
				}), nil
			} else if typ.Elem() == types[KindUnknown] {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Integer); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedType)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newIntegerDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Decoder[Value, unsafe.Pointer] {
	return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
		if s, ok := source.(Integer); ok {
			*(*T)(target) = T(s.Int())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedType)
	})
}
