package object

import (
	"encoding/binary"
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/benbjohnson/immutable"
)

// Slice is a representation of a slice.
type Slice struct {
	value *immutable.List[Object]
}

var _ Object = (*Slice)(nil)

// NewSlice returns a new Slice.
func NewSlice(values ...Object) *Slice {
	builder := immutable.NewListBuilder[Object]()
	for _, v := range values {
		builder.Append(v)
	}
	return &Slice{value: builder.List()}
}

func (s *Slice) Prepend(value Object) *Slice {
	return &Slice{value: s.value.Prepend(value)}
}

func (s *Slice) Append(value Object) *Slice {
	return &Slice{value: s.value.Append(value)}
}

func (s *Slice) Sub(start, end int) *Slice {
	return &Slice{value: s.value.Slice(start, end)}
}

func (s *Slice) Get(index int) Object {
	if index >= s.value.Len() {
		return nil
	}
	return s.value.Get(index)
}

func (s *Slice) Set(index int, value Object) *Slice {
	if index < 0 || index >= s.value.Len() {
		return s
	}
	return &Slice{value: s.value.Set(index, value)}
}

func (s *Slice) Values() []Object {
	values := make([]Object, s.value.Len())

	itr := s.value.Iterator()
	for i := 0; !itr.Done(); i++ {
		_, v := itr.Next()

		if v != nil {
			values[i] = v
		}
	}

	return values
}

func (s *Slice) Len() int {
	return s.value.Len()
}

// Slice returns a raw representation.
func (s *Slice) Slice() []any {
	values := make([]any, s.value.Len())

	itr := s.value.Iterator()
	for i := 0; !itr.Done(); i++ {
		_, v := itr.Next()

		if v != nil {
			values[i] = v.Interface()
		}
	}

	return values
}

func (s *Slice) Kind() Kind {
	return KindSlice
}

func (s *Slice) Compare(v Object) int {
	if r, ok := v.(*Slice); ok {
		minLen := s.Len()
		if minLen > r.Len() {
			minLen = r.Len()
		}

		for i := 0; i < minLen; i++ {
			if diff := Compare(s.Get(i), r.Get(i)); diff != 0 {
				return diff
			}
		}

		if s.Len() < r.Len() {
			return -1
		} else if s.Len() > r.Len() {
			return 1
		}

		return 0
	}

	if KindOf(s) > KindOf(v) {
		return 1
	}
	return -1
}

func (s *Slice) Hash() uint64 {
	h := fnv.New64a()

	var buf [8]byte
	itr := s.value.Iterator()
	for i := 0; !itr.Done(); i++ {
		_, v := itr.Next()

		_, _ = h.Write([]byte{byte(KindOf(v))})
		binary.BigEndian.PutUint64(buf[:], Hash(v))
		_, _ = h.Write(buf[:])
	}

	return h.Sum64()
}

func (s *Slice) Interface() any {
	values := make([]any, s.value.Len())

	itr := s.value.Iterator()
	for i := 0; !itr.Done(); i++ {
		_, v := itr.Next()

		if v != nil {
			values[i] = v.Interface()
		}
	}

	elementType := getCommonType(values)

	t := reflect.MakeSlice(reflect.SliceOf(elementType), s.value.Len(), s.value.Len())
	for i, value := range values {
		if value != nil {
			t.Index(i).Set(reflect.ValueOf(value))
		}
	}

	return t.Interface()
}

func newSliceEncoder(encoder *encoding.Assembler[*Object, any]) encoding.Compiler[*Object] {
	return encoding.CompilerFunc[*Object](func(typ reflect.Type) (encoding.Encoder[*Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Array || typ.Elem().Kind() == reflect.Slice {
				return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
					t := reflect.NewAt(typ.Elem(), target).Elem()

					values := make([]Object, t.Len())
					for i := 0; i < t.Len(); i++ {
						if err := encoder.Encode(&values[i], t.Index(i).Interface()); err != nil {
							return err
						}
					}

					*source = NewSlice(values...)
					return nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newSliceDecoder(decoder *encoding.Assembler[Object, any]) encoding.Compiler[Object] {
	setElement := func(source Object, target reflect.Value, i int) error {
		v := reflect.New(target.Type().Elem())
		if err := decoder.Encode(source, v.Interface()); err != nil {
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

	return encoding.CompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Array || typ.Elem().Kind() == reflect.Slice {
				return encoding.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
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
				return encoding.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
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
