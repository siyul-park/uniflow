package object

import (
	"encoding/binary"
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/benbjohnson/immutable"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Slice represents a slice of Objects.
type Slice struct {
	value *immutable.List[Object]
	hash  uint64
}

var _ Object = (*Slice)(nil)

// NewSlice returns a new Slice.
func NewSlice(elements ...Object) *Slice {
	return newSlice(immutable.NewList(elements...))
}

func newSlice(value *immutable.List[Object]) *Slice {
	h := fnv.New64a()
	var buf [8]byte
	for itr := value.Iterator(); !itr.Done(); {
		_, v := itr.Next()

		_, _ = h.Write([]byte{byte(KindOf(v))})
		binary.BigEndian.PutUint64(buf[:], HashOf(v))
		_, _ = h.Write(buf[:])
	}

	return &Slice{
		value: value,
		hash:  h.Sum64(),
	}
}

// Prepend adds a value to the beginning of the slice.
func (s *Slice) Prepend(value Object) *Slice {
	return newSlice(s.value.Prepend(value))
}

// Append adds a value to the end of the slice.
func (s *Slice) Append(value Object) *Slice {
	return newSlice(s.value.Append(value))
}

// Sub returns a new slice that is a sub-slice of the original slice.
func (s *Slice) Sub(start, end int) *Slice {
	return newSlice(s.value.Slice(start, end))
}

// Get retrieves the value at the given index.
func (s *Slice) Get(index int) Object {
	if index >= s.value.Len() {
		return nil
	}
	return s.value.Get(index)
}

// Set sets the value at the given index.
func (s *Slice) Set(index int, value Object) *Slice {
	if index < 0 || index >= s.value.Len() {
		return s
	}
	return newSlice(s.value.Set(index, value))
}

// Values returns the elements of the slice.
func (s *Slice) Values() []Object {
	elements := make([]Object, s.value.Len())
	for itr := s.value.Iterator(); !itr.Done(); {
		i, v := itr.Next()
		elements[i] = v
	}
	return elements
}

// Len returns the length of the slice.
func (s *Slice) Len() int {
	return s.value.Len()
}

// Slice returns a raw representation of the slice.
func (s *Slice) Slice() []any {
	values := make([]any, s.value.Len())
	for itr := s.value.Iterator(); !itr.Done(); {
		i, v := itr.Next()
		values[i] = InterfaceOf(v)
	}
	return values
}

// Kind returns the kind of the slice.
func (s *Slice) Kind() Kind {
	return KindSlice
}

// Equal checks if two Slice instances are equal.
func (s *Slice) Equal(other Object) bool {
	if o, ok := other.(*Slice); ok {
		if s.hash == o.hash {
			if s.value.Len() == o.value.Len() {
				itr1 := s.value.Iterator()
				itr2 := o.value.Iterator()
				for !itr1.Done() && !itr2.Done() {
					_, v1 := itr1.Next()
					_, v2 := itr2.Next()

					if !Equal(v1, v2) {
						return false
					}
				}
				return true
			}
		}
	}
	return false
}

// Compare checks whether another Object is equal to this Slice instance.
func (s *Slice) Compare(other Object) int {
	if o, ok := other.(*Slice); ok {
		itr1 := s.value.Iterator()
		itr2 := o.value.Iterator()
		for !itr1.Done() && !itr2.Done() {
			_, v1 := itr1.Next()
			_, v2 := itr2.Next()

			if c := Compare(v1, v2); c != 0 {
				return c
			}
		}
		return compare(s.value.Len(), o.value.Len())
	}
	return compare(s.Kind(), KindOf(other))
}

// Hash returns the hash value of the slice.
func (s *Slice) Hash() uint64 {
	return s.hash
}

// Interface returns the slice as a generic interface.
func (s *Slice) Interface() any {
	elements := s.Slice()
	elementType := getCommonType(elements)
	t := reflect.MakeSlice(reflect.SliceOf(elementType), len(elements), len(elements))
	for i, value := range elements {
		if value != nil {
			t.Index(i).Set(reflect.ValueOf(value))
		}
	}
	return t.Interface()
}

func newSliceEncoder(encoder *encoding.EncodeAssembler[any, Object]) encoding.EncodeCompiler[Object] {
	return encoding.EncodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[unsafe.Pointer, Object], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Array || typ.Elem().Kind() == reflect.Slice {
				valueType := reflect.PointerTo(typ.Elem().Elem())
				enc, _ := encoder.Compile(valueType)

				return encoding.EncodeFunc[unsafe.Pointer, Object](func(source unsafe.Pointer) (Object, error) {
					t := reflect.NewAt(typ.Elem(), source).Elem()

					values := make([]Object, 0, t.Len())
					for i := 0; i < t.Len(); i++ {
						v := t.Index(i)
						v = reflect.ValueOf(v.Interface())

						vPtr := reflect.New(v.Type())
						vPtr.Elem().Set(v)

						var value Object
						var err error
						if enc != nil && v.Type() == valueType.Elem() {
							value, err = enc.Encode(vPtr.UnsafePointer())
						} else {
							value, err = encoder.Encode(vPtr.Interface())
						}

						if err != nil {
							return nil, err
						} else {
							values = append(values, value)
						}
					}

					return NewSlice(values...), nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newSliceDecoder(decoder *encoding.DecodeAssembler[Object, any]) encoding.DecodeCompiler[Object] {
	setElement := func(source Object, target reflect.Value, i int) error {
		v := reflect.New(target.Type().Elem())
		if err := decoder.Decode(source, v.Interface()); err != nil {
			return err
		}

		if target.Len() < i+1 {
			if target.Kind() == reflect.Slice {
				target.Set(reflect.Append(target, v.Elem()).Convert(target.Type()))
			} else {
				return errors.WithStack(encoding.ErrInvalidValue)
			}
		} else {
			target.Index(i).Set(v.Elem().Convert(target.Type().Elem()))
		}
		return nil
	}

	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Array || typ.Elem().Kind() == reflect.Slice {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					t := reflect.NewAt(typ.Elem(), target).Elem()
					if s, ok := source.(*Slice); ok {
						for i := 0; i < s.Len(); i++ {
							if err := setElement(s.Get(i), t, i); err != nil {
								return err
							}
						}
						return nil
					}
					return setElement(source, t, 0)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(*Slice); ok {
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
