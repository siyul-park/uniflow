package primitive

import (
	"fmt"
	"reflect"

	"github.com/benbjohnson/immutable"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Slice is a representation of a slice.
type Slice struct {
	value *immutable.List[Value]
}

var _ Value = (*Slice)(nil)

// NewSlice returns a new Slice.
func NewSlice(values ...Value) *Slice {
	builder := immutable.NewListBuilder[Value]()
	for _, v := range values {
		builder.Append(v)
	}
	return &Slice{value: builder.List()}
}

func (s *Slice) Prepend(value Value) *Slice {
	return &Slice{value: s.value.Prepend(value)}
}

func (s *Slice) Append(value Value) *Slice {
	return &Slice{value: s.value.Append(value)}
}

func (s *Slice) Sub(start, end int) *Slice {
	return &Slice{value: s.value.Slice(start, end)}
}

func (s *Slice) Get(index int) Value {
	if index >= s.value.Len() {
		return nil
	}
	return s.value.Get(index)
}

func (s *Slice) Set(index int, value Value) *Slice {
	if index < 0 || index >= s.value.Len() {
		return s
	}
	return &Slice{value: s.value.Set(index, value)}
}

func (s *Slice) Len() int {
	return s.value.Len()
}

// Slice returns a raw representation.
func (s *Slice) Slice() []any {
	rawSlice := make([]any, s.value.Len())

	itr := s.value.Iterator()
	for i := 0; !itr.Done(); i++ {
		_, v := itr.Next()

		if v != nil {
			rawSlice[i] = v.Interface()
		}
	}

	return rawSlice
}

func (s *Slice) Kind() Kind {
	return KindSlice
}

func (s *Slice) Compare(v Value) int {
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

	if s.Kind() > v.Kind() {
		return 1
	}
	return -1
}

func (s *Slice) Interface() any {
	var values []any

	itr := s.value.Iterator()
	for i := 0; !itr.Done(); i++ {
		_, v := itr.Next()

		if v != nil {
			values = append(values, v.Interface())
		} else {
			values = append(values, nil)
		}
	}

	elementType := getCommonType(values)

	sliceValue := reflect.MakeSlice(reflect.SliceOf(elementType), s.value.Len(), s.value.Len())
	for i, value := range values {
		if value != nil {
			sliceValue.Index(i).Set(reflect.ValueOf(value))
		}
	}

	return sliceValue.Interface()
}

func newSliceEncoder(encoder encoding.Encoder[any, Value]) encoding.Encoder[any, Value] {
	return encoding.EncoderFunc[any, Value](func(source any) (Value, error) {
		if s := reflect.ValueOf(source); s.Kind() == reflect.Slice || s.Kind() == reflect.Array {
			values := make([]Value, s.Len())
			for i := 0; i < s.Len(); i++ {
				if v, err := encoder.Encode(s.Index(i).Interface()); err != nil {
					return nil, err
				} else {
					values[i] = v
				}
			}
			return NewSlice(values...), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newSliceDecoder(decoder encoding.Decoder[Value, any]) encoding.Decoder[Value, any] {
	return encoding.DecoderFunc[Value, any](func(source Value, target any) error {
		if s, ok := source.(*Slice); ok {
			if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				if t.Elem().Kind() == reflect.Slice || t.Elem().Kind() == reflect.Array {
					for i := 0; i < s.Len(); i++ {
						value := s.Get(i)
						v := reflect.New(t.Elem().Type().Elem())
						if err := decoder.Decode(value, v.Interface()); err != nil {
							return errors.WithMessage(err, fmt.Sprintf("value(%v) corresponding to the index(%v) cannot be decoded", value.Interface(), i))
						}
						if t.Elem().Len() < i+1 {
							if t.Elem().Kind() == reflect.Slice {
								t.Elem().Set(reflect.Append(t.Elem(), v.Elem()))
							} else {
								return errors.WithMessage(encoding.ErrUnsupportedValue, fmt.Sprintf("index(%d) is exceeded len(%d)", i, t.Elem().Len()))
							}
						} else {
							t.Elem().Index(i).Set(v.Elem())
						}
					}
					return nil
				} else if t.Elem().Type() == typeAny {
					t.Elem().Set(reflect.ValueOf(s.Interface()))
					return nil
				}
			}
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
