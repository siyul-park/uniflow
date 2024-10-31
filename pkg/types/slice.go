package types

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
type Slice = *slice_

type slice_ struct {
	value *immutable.List[Value]
}

var _ Value = (Slice)(nil)

// NewSlice returns a new Slice.
func NewSlice(elements ...Value) Slice {
	return &slice_{value: immutable.NewList(elements...)}
}

// Prepend adds a value to the beginning of the slice.
func (s Slice) Prepend(value Value) Slice {
	return &slice_{value: s.value.Prepend(value)}
}

// Append adds a value to the end of the slice.
func (s Slice) Append(value Value) Slice {
	return &slice_{value: s.value.Append(value)}
}

// Sub returns a new slice that is a sub-slice of the original slice.
func (s Slice) Sub(start, end int) Slice {
	return &slice_{value: s.value.Slice(start, end)}
}

// Get retrieves the value at the given index.
func (s Slice) Get(index int) Value {
	if index >= s.value.Len() {
		return nil
	}
	return s.value.Get(index)
}

// Set sets the value at the given index.
func (s Slice) Set(index int, value Value) Slice {
	if index < 0 || index >= s.value.Len() {
		return s
	}
	return &slice_{value: s.value.Set(index, value)}
}

// Values returns the elements of the slice.
func (s Slice) Values() []Value {
	elements := make([]Value, s.value.Len())
	for itr := s.value.Iterator(); !itr.Done(); {
		i, v := itr.Next()
		elements[i] = v
	}
	return elements
}

// Len returns the length of the slice.
func (s Slice) Len() int {
	return s.value.Len()
}

// Slice returns a raw representation of the slice.
func (s Slice) Slice() []any {
	if s.value.Len() == 0 {
		return nil
	}

	values := make([]any, s.value.Len())
	for itr := s.value.Iterator(); !itr.Done(); {
		i, v := itr.Next()
		values[i] = InterfaceOf(v)
	}

	return values
}

// Kind returns the kind of the slice.
func (s Slice) Kind() Kind {
	return KindSlice
}

// Hash returns the hash value of the slice.
func (s Slice) Hash() uint64 {
	h := fnv.New64a()
	var buf [8]byte
	for itr := s.value.Iterator(); !itr.Done(); {
		_, v := itr.Next()

		binary.BigEndian.PutUint64(buf[:], HashOf(v))
		_, _ = h.Write(buf[:])
	}
	return h.Sum64()
}

// Interface returns the slice as a generic interface.
func (s Slice) Interface() any {
	if s.value.Len() == 0 {
		return nil
	}

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

// Equal checks if two Slice instances are equal.
func (s Slice) Equal(other Value) bool {
	if o, ok := other.(Slice); ok {
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
	return false
}

// Compare checks whether another Object is equal to this Slice instance.
func (s Slice) Compare(other Value) int {
	if o, ok := other.(Slice); ok {
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

func newSliceEncoder(encoder *encoding.EncodeAssembler[any, Value]) encoding.EncodeCompiler[any, Value] {
	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ != nil && (typ.Kind() == reflect.Array || typ.Kind() == reflect.Slice) {
			valueEncoder, _ := encoder.Compile(typ.Elem())
			if valueEncoder == nil {
				valueEncoder = encoder
			}

			return encoding.EncodeFunc(func(source any) (Value, error) {
				s := reflect.ValueOf(source)

				values := make([]Value, 0, s.Len())
				for i := 0; i < s.Len(); i++ {
					v := s.Index(i)

					if value, err := valueEncoder.Encode(v.Interface()); err != nil {
						return nil, err
					} else {
						values = append(values, value)
					}
				}
				return NewSlice(values...), nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newSliceDecoder(decoder *encoding.DecodeAssembler[Value, any]) encoding.DecodeCompiler[Value] {
	setElement := func(source Value, target reflect.Value, i int) error {
		v := reflect.New(target.Type().Elem())
		if err := decoder.Decode(source, v.Interface()); err != nil {
			return err
		}

		if target.Len() < i+1 {
			if target.Kind() != reflect.Slice {
				return errors.WithStack(encoding.ErrUnsupportedValue)
			} else {
				target.Set(reflect.Append(target, v.Elem()).Convert(target.Type()))
			}
		} else {
			target.Index(i).Set(v.Elem().Convert(target.Type().Elem()))
		}
		return nil
	}

	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Array || typ.Elem().Kind() == reflect.Slice {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					t := reflect.NewAt(typ.Elem(), target).Elem()
					if s, ok := source.(Slice); ok {
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
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Slice); ok {
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
