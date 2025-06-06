package types

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"

	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Uinteger is an interface representing an unsigned integer.
type Uinteger interface {
	Value
	Uint() uint64
}

// Uint is an interface representing an unsigned integer.
type Uint struct {
	value uint
}

// Uint8 represents a uint8 type.
type Uint8 struct {
	value uint8
}

// Uint16 represents a uint16 type.
type Uint16 struct {
	value uint16
}

// Uint32 represents a uint32 type.
type Uint32 struct {
	value uint32
}

// Uint64 represents a uint64 type.
type Uint64 struct {
	value uint64
}

var (
	_ Uinteger = Uint{}
	_ Uinteger = Uint8{}
	_ Uinteger = Uint16{}
	_ Uinteger = Uint32{}
	_ Uinteger = Uint64{}

	_ json.Marshaler = Uint{}
	_ json.Marshaler = Uint8{}
	_ json.Marshaler = Uint16{}
	_ json.Marshaler = Uint32{}
	_ json.Marshaler = Uint64{}

	_ json.Unmarshaler = (*Uint)(nil)
	_ json.Unmarshaler = (*Uint8)(nil)
	_ json.Unmarshaler = (*Uint16)(nil)
	_ json.Unmarshaler = (*Uint32)(nil)
	_ json.Unmarshaler = (*Uint64)(nil)
)

// NewUint returns a new Uint instance.
func NewUint(value uint) Uint {
	return Uint{value: value}
}

// Uint returns the raw representation of the unsigned integer.
func (u Uint) Uint() uint64 {
	return uint64(u.value)
}

// Kind returns the type of the unsigned integer data.
func (u Uint) Kind() Kind {
	return KindUint
}

// Hash returns the hash code for the unsigned integer value.
func (u Uint) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[unsafe.Sizeof(u.value)]byte)(unsafe.Pointer(&u.value))[:])
	return h.Sum64()
}

// Interface converts Uint to a uint64.
func (u Uint) Interface() any {
	return u.value
}

// Equal checks if two Uint typess are equal.
func (u Uint) Equal(other Value) bool {
	if o, ok := other.(Uint); ok {
		return u.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Uint instance.
func (u Uint) Compare(other Value) int {
	if o, ok := other.(Uint); ok {
		return compare(u.value, o.value)
	}
	return compare(u.Kind(), KindOf(other))
}

// MarshalJSON implements the encoding.MarshalJSON interface.
func (i Uint) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// UnmarshalJSON implements the encoding.UnmarshalJSON interface.
func (i *Uint) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &i.value)
}

// NewUint8 returns a new Uint8 instance.
func NewUint8(value uint8) Uint8 {
	return Uint8{value: value}
}

// Uint returns the raw representation of the unsigned integer.
func (u Uint8) Uint() uint64 {
	return uint64(u.value)
}

// kind returns the type of the unsigned integer data.
func (u Uint8) Kind() Kind {
	return KindUint8
}

// Hash returns the hash code for the unsigned integer value.
func (u Uint8) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[1]byte)(unsafe.Pointer(&u.value))[:])
	return h.Sum64()
}

// Interface converts Uint8 to a uint8.
func (u Uint8) Interface() any {
	return u.value
}

// Equal checks if two Uint8 typess are equal.
func (u Uint8) Equal(other Value) bool {
	if o, ok := other.(Uint8); ok {
		return u.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Uint8 instance.
func (u Uint8) Compare(other Value) int {
	if o, ok := other.(Uint8); ok {
		return compare(u.value, o.value)
	}
	return compare(u.Kind(), KindOf(other))
}

// MarshalJSON implements the encoding.MarshalJSON interface.
func (i Uint8) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// UnmarshalJSON implements the encoding.UnmarshalJSON interface.
func (i *Uint8) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &i.value)
}

// NewUint16 returns a new Uint16 instance.
func NewUint16(value uint16) Uint16 {
	return Uint16{value: value}
}

// Uint returns the raw representation of the unsigned integer.
func (u Uint16) Uint() uint64 {
	return uint64(u.value)
}

// kind returns the type of the unsigned integer data.
func (u Uint16) Kind() Kind {
	return KindUint16
}

// Hash returns the hash code for the unsigned integer value.
func (u Uint16) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[2]byte)(unsafe.Pointer(&u.value))[:])
	return h.Sum64()
}

// Interface converts Uint16 to a uint16.
func (u Uint16) Interface() any {
	return u.value
}

// Equal checks if two Uint16 typess are equal.
func (u Uint16) Equal(other Value) bool {
	if o, ok := other.(Uint16); ok {
		return u.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Uint16 instance.
func (u Uint16) Compare(other Value) int {
	if o, ok := other.(Uint16); ok {
		return compare(u.value, o.value)
	}
	return compare(u.Kind(), KindOf(other))
}

// MarshalJSON implements the encoding.MarshalJSON interface.
func (i Uint16) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// UnmarshalJSON implements the encoding.UnmarshalJSON interface.
func (i *Uint16) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &i.value)
}

// NewUint32 returns a new Uint32 instance.
func NewUint32(value uint32) Uint32 {
	return Uint32{value: value}
}

// Uint returns the raw representation of the unsigned integer.
func (u Uint32) Uint() uint64 {
	return uint64(u.value)
}

// kind returns the type of the unsigned integer data.
func (u Uint32) Kind() Kind {
	return KindUint32
}

// Hash returns the hash code for the unsigned integer value.
func (u Uint32) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[4]byte)(unsafe.Pointer(&u.value))[:])
	return h.Sum64()
}

// Interface converts Uint32 to a uint32.
func (u Uint32) Interface() any {
	return u.value
}

// Equal checks if two Uint32 typess are equal.
func (u Uint32) Equal(other Value) bool {
	if o, ok := other.(Uint32); ok {
		return u.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Uint32 instance.
func (u Uint32) Compare(other Value) int {
	if o, ok := other.(Uint32); ok {
		return compare(u.value, o.value)
	}
	return compare(u.Kind(), KindOf(other))
}

// MarshalJSON implements the encoding.MarshalJSON interface.
func (i Uint32) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// UnmarshalJSON implements the encoding.UnmarshalJSON interface.
func (i *Uint32) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &i.value)
}

// NewUint64 returns a new Uint64 instance.
func NewUint64(value uint64) Uint64 {
	return Uint64{value: value}
}

// Uint returns the raw representation of the unsigned integer.
func (u Uint64) Uint() uint64 {
	return u.value
}

// Kind returns the type of the unsigned integer data.
func (u Uint64) Kind() Kind {
	return KindUint64
}

// Hash returns the hash code for the unsigned integer value.
func (u Uint64) Hash() uint64 {
	h := fnv.New64a()
	h.Write((*[8]byte)(unsafe.Pointer(&u.value))[:])
	return h.Sum64()
}

// Interface converts Uint64 to a uint64.
func (u Uint64) Interface() any {
	return u.value
}

// Equal checks if two Uint64 typess are equal.
func (u Uint64) Equal(other Value) bool {
	if o, ok := other.(Uint64); ok {
		return u.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Uint64 instance.
func (u Uint64) Compare(other Value) int {
	if o, ok := other.(Uint64); ok {
		return compare(u.value, o.value)
	}
	return compare(u.Kind(), KindOf(other))
}

// MarshalJSON implements the encoding.MarshalJSON interface.
func (i Uint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

// UnmarshalJSON implements the encoding.UnmarshalJSON interface.
func (i *Uint64) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &i.value)
}

func newUintegerEncoder() encoding.EncodeCompiler[any, Value] {
	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ == nil {
			return nil, errors.WithStack(encoding.ErrUnsupportedType)
		} else if typ.Kind() == reflect.Uint {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				if s, ok := source.(uint); ok {
					return NewUint(s), nil
				} else {
					return NewUint(uint(reflect.ValueOf(source).Uint())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Uint8 {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				if s, ok := source.(uint8); ok {
					return NewUint8(s), nil
				} else {
					return NewUint8(uint8(reflect.ValueOf(source).Uint())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Uint16 {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				if s, ok := source.(uint16); ok {
					return NewUint16(s), nil
				} else {
					return NewUint16(uint16(reflect.ValueOf(source).Uint())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Uint32 {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				if s, ok := source.(uint32); ok {
					return NewUint32(s), nil
				} else {
					return NewUint32(uint32(reflect.ValueOf(source).Uint())), nil
				}
			}), nil
		} else if typ.Kind() == reflect.Uint64 {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				if s, ok := source.(uint64); ok {
					return NewUint64(s), nil
				} else {
					return NewUint64(reflect.ValueOf(source).Uint()), nil
				}
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newUintegerDecoder() encoding.DecodeCompiler[Value] {
	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
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
			} else if typ.Elem().Kind() == reflect.String {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Uinteger); ok {
						*(*string)(target) = fmt.Sprint(s.Interface())
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedType)
				}), nil
			} else if typ.Elem() == types[KindUnknown] {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Uinteger); ok {
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

func newUintegerDecoderWithType[T constraints.Integer | constraints.Float]() encoding.Decoder[Value, unsafe.Pointer] {
	return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
		if s, ok := source.(Uinteger); ok {
			*(*T)(target) = T(s.Uint())
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedType)
	})
}
