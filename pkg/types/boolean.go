package types

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Boolean is a representation of a boolean value.
type Boolean struct {
	value bool
}

var _ Value = Boolean{}

// Predefined True and False values for optimization.
var (
	True  = Boolean{value: true}
	False = Boolean{value: false}
)

// NewBoolean returns the predefined True or False instance.
func NewBoolean(value bool) Boolean {
	if value {
		return True
	}
	return False
}

// Bool returns the raw boolean value.
func (b Boolean) Bool() bool {
	return b.value
}

// Kind returns the kind of the boolean data.
func (b Boolean) Kind() Kind {
	return KindBoolean
}

// Hash returns the hash code for the boolean value.
func (b Boolean) Hash() uint64 {
	h := fnv.New64a()
	var value byte
	if b.value {
		value = 1
	}
	h.Write([]byte{value})
	return h.Sum64()
}

// Interface converts Boolean to a generic interface.
func (b Boolean) Interface() any {
	return b.value
}

// Equal checks if the other Object is equal to this Boolean.
func (b Boolean) Equal(other Value) bool {
	if o, ok := other.(Boolean); ok {
		return b.value == o.value
	}
	return false
}

// Compare compares another Object with this Boolean instance.
func (b Boolean) Compare(other Value) int {
	if o, ok := other.(Boolean); ok {
		if b.value == o.value {
			return 0
		}
		if b.value {
			return 1
		}
		return -1
	}
	return compare(b.Kind(), KindOf(other))
}

func newBooleanEncoder() encoding.EncodeCompiler[any, Value] {
	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ != nil && typ.Kind() == reflect.Bool {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				if s, ok := source.(bool); ok {
					return NewBoolean(s), nil
				} else {
					return NewBoolean(reflect.ValueOf(source).Bool()), nil
				}
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newBooleanDecoder() encoding.DecodeCompiler[Value] {
	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Bool {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Boolean); ok {
						*(*bool)(target) = s.Bool()
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedType)
				}), nil
			} else if typ.Elem().Kind() == reflect.String {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Boolean); ok {
						*(*string)(target) = fmt.Sprint(s.Interface())
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedType)
				}), nil
			} else if typ.Elem() == types[KindUnknown] {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Boolean); ok {
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
