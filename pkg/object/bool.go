package object

import (
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Bool is a representation of a boolean value.
type Bool struct {
	value bool
}

var _ Object = Bool{}

// Predefined True and False values for optimization.
var (
	True  = Bool{value: true}
	False = Bool{value: false}
)

// NewBool returns a pointer to a Bool instance.
func NewBool(value bool) Bool {
	if value {
		return True
	}
	return False
}

// Bool returns the raw boolean value.
func (b Bool) Bool() bool {
	return b.value
}

// Kind returns the kind of the boolean data.
func (b Bool) Kind() Kind {
	return KindBool
}

// Hash returns the hash code for the boolean value.
func (b Bool) Hash() uint64 {
	h := fnv.New64a()
	var value byte
	if b.value {
		value = 1
	} else {
		value = 0
	}
	h.Write([]byte{value})
	return h.Sum64()
}

// Interface converts Bool to a generic interface.
func (b Bool) Interface() any {
	return b.value
}

// Equal checks if the other Object is equal to this Bool.
func (b Bool) Equal(other Object) bool {
	if o, ok := other.(Bool); ok {
		return b.value == o.value
	}
	return false
}

// Compare checks whether another Object is equal to this Bool instance.
func (b Bool) Compare(other Object) int {
	if o, ok := other.(Bool); ok {
		if b.value == o.value {
			return 0
		} else if !b.value && o.value {
			return -1
		} else {
			return 1
		}
	}
	return compare(b.Kind(), KindOf(other))
}

func newBoolEncoder() encoding.EncodeCompiler[any, Object] {
	return encoding.EncodeCompilerFunc[any, Object](func(typ reflect.Type) (encoding.Encoder[any, Object], error) {
		if typ.Kind() == reflect.Bool {
			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				if s, ok := source.(bool); ok {
					return NewBool(s), nil
				} else {
					return NewBool(reflect.ValueOf(source).Bool()), nil
				}
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newBoolDecoder() encoding.DecodeCompiler[Object] {
	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Bool {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(Bool); ok {
						*(*bool)(target) = s.Bool()
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(Bool); ok {
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
