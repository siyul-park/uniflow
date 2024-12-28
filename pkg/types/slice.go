package types

import (
	"encoding/binary"
	"hash/fnv"
	"reflect"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Slice represents a slice of Objects.
type Slice = *_slice

type _slice struct {
	value []Value
	hash  uint64
	mu    sync.RWMutex
}

var _ Value = (Slice)(nil)

// NewSlice returns a new Slice.
func NewSlice(elements ...Value) Slice {
	return &_slice{value: elements}
}

// Prepend adds a value to the beginning of the slice.
func (s Slice) Prepend(elements ...Value) Slice {
	return &_slice{value: append(elements, s.value...)}
}

// Append adds a value to the end of the slice.
func (s Slice) Append(elements ...Value) Slice {
	value := make([]Value, len(s.value), len(s.value)+len(elements))
	copy(value, s.value)
	value = append(value, elements...)

	return &_slice{value: value}
}

// Sub returns a new slice that is a sub-slice of the original slice.
func (s Slice) Sub(start, end int) Slice {
	if start < 0 {
		start = 0
	}
	if end > len(s.value) {
		end = len(s.value)
	}
	if end <= start {
		return &_slice{}
	}

	elements := make([]Value, end-start)
	copy(elements, s.value[start:end])

	return &_slice{value: elements}
}

// Get retrieves the value at the given index.
func (s Slice) Get(index int) Value {
	if index >= len(s.value) {
		return nil
	}
	return s.value[index]
}

// Set sets the value at the given index.
func (s Slice) Set(index int, value Value) Slice {
	if index < 0 || index >= len(s.value) {
		return s
	}

	elements := make([]Value, len(s.value))
	copy(elements, s.value)
	elements[index] = value

	return &_slice{value: elements}
}

// Values returns the elements of the slice.
func (s Slice) Values() []Value {
	return append([]Value(nil), s.value...)
}

// Range returns a function that iterates over all key-value pairs of the slice.
func (s Slice) Range() func(func(key int, value Value) bool) {
	return func(yield func(key int, value Value) bool) {
		for i := 0; i < len(s.value); i++ {
			v := s.value[i]
			if !yield(i, v) {
				return
			}
		}
	}
}

// Len returns the length of the slice.
func (s Slice) Len() int {
	return len(s.value)
}

// Slice returns a raw representation of the slice.
func (s Slice) Slice() []any {
	if len(s.value) == 0 {
		return nil
	}

	values := make([]any, len(s.value))
	for i := 0; i < len(s.value); i++ {
		v := s.value[i]
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.hash == 0 {
		h := fnv.New64a()
		var buf [8]byte
		for i := 0; i < len(s.value); i++ {
			v := s.value[i]
			binary.BigEndian.PutUint64(buf[:], HashOf(v))
			_, _ = h.Write(buf[:])
		}
		s.hash = h.Sum64()
	}
	return s.hash
}

// Interface returns the slice as a generic interface.
func (s Slice) Interface() any {
	if len(s.value) == 0 {
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
		if s.Hash() != o.Hash() {
			return false
		}

		if len(s.value) == len(o.value) {
			for i := 0; i < len(s.value); i++ {
				v1 := s.value[i]
				v2 := o.value[i]

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
		length := min(len(s.value), len(o.value))
		for i := 0; i < length; i++ {
			v1 := s.value[i]
			v2 := o.value[i]

			if c := Compare(v1, v2); c != 0 {
				return c
			}
		}
		return compare(len(s.value), len(o.value))
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
	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Array || typ.Elem().Kind() == reflect.Slice {
				valueDecoder, err := decoder.Compile(reflect.PointerTo(typ.Elem().Elem()))
				if err != nil {
					return nil, err
				}

				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					t := reflect.NewAt(typ.Elem(), target).Elem()
					if s, ok := source.(Slice); ok {
						for t.Len() < s.Len() {
							if t.Kind() != reflect.Slice {
								return errors.WithStack(encoding.ErrUnsupportedValue)
							} else {
								t.Set(reflect.Append(t, reflect.Zero(t.Type().Elem())))
							}
						}

						for i, v := range s.Range() {
							if err := valueDecoder.Decode(v, t.Index(i).Addr().UnsafePointer()); err != nil {
								return err
							}
						}
						return nil
					}

					if t.Len() == 0 {
						if t.Kind() != reflect.Slice {
							return errors.WithStack(encoding.ErrUnsupportedValue)
						} else {
							t.Set(reflect.Append(t, reflect.Zero(t.Type().Elem())))
						}
					}
					return valueDecoder.Decode(source, t.Index(0).Addr().UnsafePointer())
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
