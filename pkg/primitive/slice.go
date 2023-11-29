package primitive

import (
	"fmt"
	"reflect"

	"github.com/benbjohnson/immutable"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

type (
	// Slice is a representation of a slice.
	Slice struct {
		value *immutable.List[Value]
	}
)

var _ Value = (*Slice)(nil)

// NewSlice returns a new Slice.
func NewSlice(values ...Value) *Slice {
	b := immutable.NewListBuilder[Value]()
	for _, v := range values {
		b.Append(v)
	}
	return &Slice{value: b.List()}
}

func (o *Slice) Prepend(value Value) *Slice {
	return &Slice{value: o.value.Prepend(value)}
}

func (o *Slice) Append(value Value) *Slice {
	return &Slice{value: o.value.Append(value)}
}

func (o *Slice) Sub(start, end int) *Slice {
	return &Slice{value: o.value.Slice(start, end)}
}

func (o *Slice) Get(index int) Value {
	if index >= o.value.Len() {
		return nil
	}
	return o.value.Get(index)
}

func (o *Slice) Set(index int, value Value) *Slice {
	if index < 0 && index >= o.value.Len() {
		return o
	}
	return &Slice{value: o.value.Set(index, value)}
}

func (o *Slice) Len() int {
	return o.value.Len()
}

// Slice returns a raw representation.
func (o *Slice) Slice() []any {
	// TODO: support more type defined slice.
	s := make([]any, o.value.Len())

	itr := o.value.Iterator()
	for !itr.Done() {
		i, v := itr.Next()

		if v != nil {
			s[i] = v.Interface()
		}
	}

	return s
}

func (o *Slice) Kind() Kind {
	return KindSlice
}

func (o *Slice) Compare(v Value) int {
	if r, ok := v.(*Slice); !ok {
		if o.Kind() > v.Kind() {
			return 1
		} else {
			return -1
		}
	} else {
		for i := 0; i < o.Len(); i++ {
			if r.Len() == i {
				return 1
			}

			if diff := Compare(o.Get(i), r.Get(i)); diff != 0 {
				return diff
			}
		}

		if o.Len() > r.Len() {
			return -1
		}
		return 0
	}
}

func (o *Slice) Interface() any {
	var values []any
	itr := o.value.Iterator()
	for !itr.Done() {
		_, v := itr.Next()
		if v != nil {
			values = append(values, v.Interface())
		} else {
			values = append(values, nil)
		}
	}

	valueType := typeAny

	for i, value := range values {
		typ := reflect.TypeOf(value)
		if i == 0 {
			valueType = typ
		} else if valueType != typ {
			valueType = typeAny
		}
	}

	t := reflect.MakeSlice(reflect.SliceOf(valueType), o.value.Len(), o.value.Len())
	for i, value := range values {
		t.Index(i).Set(reflect.ValueOf(value))
	}
	return t.Interface()
}

// NewSliceEncoder is encode slice or array to Slice.
func NewSliceEncoder(encoder encoding.Encoder[any, Value]) encoding.Encoder[any, Value] {
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

// NewSliceDecoder is decode Slice to slice or array.
func NewSliceDecoder(decoder encoding.Decoder[Value, any]) encoding.Decoder[Value, any] {
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
