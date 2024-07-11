package types

import (
	"encoding"
	"hash/fnv"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/pkg/errors"
	encoding2 "github.com/siyul-park/uniflow/pkg/encoding"
)

// String represents a string.
type String struct {
	value string
}

var _ Value = String{}

// NewString creates a new String instance.
func NewString(value string) String {
	return String{value: value}
}

// Len returns the length of the string.
func (s String) Len() int {
	return len([]rune(s.value))
}

// Get returns the rune at the specified index in the string.
func (s String) Get(index int) rune {
	runes := []rune(s.value)
	if index >= len(runes) {
		return rune(0)
	}
	return runes[index]
}

// String returns the raw string representation.
func (s String) String() string {
	return s.value
}

// Kind returns the kind of the value.
func (s String) Kind() Kind {
	return KindString
}

// Hash calculates and returns the hash code using FNV-1a algorithm.
func (s String) Hash() uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s.value))
	return h.Sum64()
}

// Interface converts String to its underlying string.
func (s String) Interface() any {
	return s.value
}

// Equal checks if two String instances are equal.
func (s String) Equal(other Value) bool {
	if o, ok := other.(String); ok {
		return s.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this String instance.
func (s String) Compare(other Value) int {
	if o, ok := other.(String); ok {
		return compare(s.value, o.value)
	}
	return compare(s.Kind(), KindOf(other))
}

func newStringEncoder() encoding2.EncodeCompiler[any, Value] {
	typeTextMarshaler := reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

	return encoding2.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding2.Encoder[any, Value], error) {
		if typ != nil && typ.ConvertibleTo(typeTextMarshaler) {
			return encoding2.EncodeFunc[any, Value](func(source any) (Value, error) {
				s := source.(encoding.TextMarshaler)
				if s, err := s.MarshalText(); err != nil {
					return nil, err
				} else {
					return NewString(string(s)), nil
				}
			}), nil
		} else if typ != nil && typ.Kind() == reflect.String {
			return encoding2.EncodeFunc[any, Value](func(source any) (Value, error) {
				if s, ok := source.(string); ok {
					return NewString(s), nil
				} else {
					return NewString(reflect.ValueOf(source).String()), nil
				}
			}), nil
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedType)
	})
}

func newStringDecoder() encoding2.DecodeCompiler[Value] {
	typeTextUnmarshaler := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeBinaryUnmarshaler := reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()

	return encoding2.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding2.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.ConvertibleTo(typeTextUnmarshaler) {
			return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				if s, ok := source.(String); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.TextUnmarshaler)
					return t.UnmarshalText([]byte(s.String()))
				}
				return errors.WithStack(encoding2.ErrUnsupportedType)
			}), nil
		} else if typ != nil && typ.ConvertibleTo(typeBinaryUnmarshaler) {
			return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				if s, ok := source.(String); ok {
					t := reflect.NewAt(typ.Elem(), target).Interface().(encoding.BinaryUnmarshaler)
					return t.UnmarshalBinary([]byte(s.String()))
				}
				return errors.WithStack(encoding2.ErrUnsupportedType)
			}), nil
		} else if typ != nil && typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.String {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						*(*string)(target) = s.String()
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Bool {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.ParseBool(s.String()); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*bool)(target) = bool(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Float32 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.ParseFloat(s.String(), 32); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*float32)(target) = float32(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Float64 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.ParseFloat(s.String(), 64); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*float64)(target) = float64(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Int {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.Atoi(s.String()); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*int)(target) = int(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Int8 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.Atoi(s.String()); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*int8)(target) = int8(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Int16 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.Atoi(s.String()); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*int16)(target) = int16(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Int32 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.Atoi(s.String()); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*int32)(target) = int32(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Int64 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.ParseInt(s.String(), 10, 64); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*int64)(target) = int64(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.ParseUint(s.String(), 10, 64); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*uint)(target) = uint(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint8 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.ParseUint(s.String(), 10, 8); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*uint8)(target) = uint8(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint16 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.ParseUint(s.String(), 10, 16); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*uint16)(target) = uint16(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint32 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.ParseUint(s.String(), 10, 32); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*uint32)(target) = uint32(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Uint64 {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						if v, err := strconv.ParseUint(s.String(), 10, 64); err != nil {
							return errors.WithMessage(encoding2.ErrUnsupportedValue, err.Error())
						} else {
							*(*uint64)(target) = uint64(v)
							return nil
						}
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding2.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(String); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding2.ErrUnsupportedType)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding2.ErrUnsupportedType)
	})
}
