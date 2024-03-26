package primitive

import (
	"fmt"
	"github.com/benbjohnson/immutable"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

// Map represents a map structure.
type Map struct {
	value *immutable.SortedMap[Value, Value]
}

// mapTag represents the tag for map fields.
type mapTag struct {
	alias     string
	ignore    bool
	omitempty bool
	inline    bool
}

type comparer struct{}

const tagMap = "map"

var _ Value = (*Map)(nil)
var _ immutable.Comparer[Value] = (*comparer)(nil)

// NewMap creates a new Map with key-value pairs.
func NewMap(pairs ...Value) *Map {
	builder := immutable.NewSortedMapBuilder[Value, Value](&comparer{})
	for i := 0; i < len(pairs)/2; i++ {
		k, v := pairs[i*2], pairs[i*2+1]
		builder.Set(k, v)
	}
	return &Map{value: builder.Map()}
}

// Get retrieves the value for a given key.
func (m *Map) Get(key Value) (Value, bool) {
	return m.value.Get(key)
}

// GetOr returns the value for a given key or a default value if the key is not found.
func (m *Map) GetOr(key, value Value) Value {
	if v, ok := m.Get(key); ok {
		return v
	}
	return value
}

// Set adds or updates a key-value pair in the map.
func (m *Map) Set(key, value Value) *Map {
	return &Map{value: m.value.Set(key, value)}
}

// Delete removes a key and its corresponding value from the map.
func (m *Map) Delete(key Value) *Map {
	return &Map{value: m.value.Delete(key)}
}

// Keys returns all keys in the map.
func (m *Map) Keys() []Value {
	var keys []Value
	itr := m.value.Iterator()

	for !itr.Done() {
		k, _, _ := itr.Next()
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values in the map.
func (m *Map) Values() []Value {
	var values []Value
	itr := m.value.Iterator()

	for !itr.Done() {
		_, v, _ := itr.Next()
		values = append(values, v)
	}
	return values
}

// Pairs returns all keys and values in the map.
func (m *Map) Pairs() []Value {
	var pairs []Value
	itr := m.value.Iterator()

	for !itr.Done() {
		k, v, _ := itr.Next()
		pairs = append(pairs, k, v)
	}
	return pairs
}

// Len returns the number of key-value pairs in the map.
func (m *Map) Len() int {
	return m.value.Len()
}

// Map converts the Map to a raw Go map.
func (m *Map) Map() map[any]any {
	result := make(map[any]any, m.value.Len())
	itr := m.value.Iterator()

	for !itr.Done() {
		k, v, _ := itr.Next()

		if k != nil {
			result[k.Interface()] = v.Interface()
		}
	}

	return result
}

// Merge merges the contents of the other Map into the current Map.
// If there are any overlapping keys, the values from the other Map will overwrite the values in the current Map.
func (m *Map) Merge(other *Map) *Map {
	return NewMap(append(m.Pairs(), other.Pairs()...)...)
}

// Kind returns the kind of the Map.
func (m *Map) Kind() Kind {
	return KindMap
}

// Compare compares two maps.
func (m *Map) Compare(v Value) int {
	if r, ok := v.(*Map); ok {
		keys1, keys2 := m.Keys(), r.Keys()

		if len(keys1) < len(keys2) {
			return -1
		} else if len(keys1) > len(keys2) {
			return 1
		}

		for i, k1 := range keys1 {
			k2 := keys2[i]
			if diff := Compare(k1, k2); diff != 0 {
				return diff
			}

			v1, ok1 := m.Get(k1)
			v2, ok2 := r.Get(k2)

			if diff := Compare(NewBool(ok1), NewBool(ok2)); diff != 0 {
				return diff
			}
			if diff := Compare(v1, v2); diff != 0 {
				return diff
			}
		}

		return 0
	}

	if m.Kind() > v.Kind() {
		return 1
	}
	return -1
}

// Interface converts the Map to an interface{}.
func (m *Map) Interface() any {
	var keys []any
	var values []any

	itr := m.value.Iterator()

	for !itr.Done() {
		k, v, _ := itr.Next()

		if k != nil {
			keys = append(keys, k.Interface())
		} else {
			keys = append(keys, nil)
		}
		if v != nil {
			values = append(values, v.Interface())
		} else {
			values = append(values, nil)
		}
	}

	keyType := getCommonType(keys)
	valueType := getCommonType(values)

	t := reflect.MakeMapWithSize(reflect.MapOf(keyType, valueType), len(keys))
	for i, key := range keys {
		value := values[i]
		t.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}
	return t.Interface()
}

func (*comparer) Compare(a Value, b Value) int {
	return Compare(a, b)
}

func newMapEncoder(encoder encoding.Encoder[any, Value]) encoding.Encoder[any, Value] {
	typeToEncoder := sync.Map{} // map[reflect.Type]encoding.Encoder[reflect.Value, *Map]{}

	compile := func(typ reflect.Type) (encoding.Encoder[reflect.Value, *Map], error) {
		if typ.Kind() == reflect.Map {
			return encoding.EncoderFunc[reflect.Value, *Map](func(source reflect.Value) (*Map, error) {
				pairs := make([]Value, 0, len(source.MapKeys())*2)
				for _, k := range source.MapKeys() {
					if k, err := encoder.Encode(k.Interface()); err != nil {
						return nil, errors.WithMessage(err, fmt.Sprintf("key(%v) can't encode", k.Interface()))
					} else {
						pairs = append(pairs, k)
					}

					if v, err := encoder.Encode(source.MapIndex(k).Interface()); err != nil {
						return nil, errors.WithMessage(err, fmt.Sprintf("value(%v) can't encode", source.MapIndex(k).Interface()))
					} else {
						pairs = append(pairs, v)
					}
				}
				return NewMap(pairs...), nil
			}), nil
		} else if typ.Kind() == reflect.Struct {
			var encoders []encoding.Encoder[reflect.Value, []Value]
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)
				tag := getMapTag(field)

				if !field.IsExported() || tag.ignore {
					continue
				}

				var enc encoding.Encoder[reflect.Value, []Value]
				if tag.inline {
					enc = encoding.EncoderFunc[reflect.Value, []Value](func(source reflect.Value) ([]Value, error) {
						if target, err := encoder.Encode(source.Field(i).Interface()); err != nil {
							return nil, err
						} else if t, ok := target.(*Map); !ok {
							return nil, errors.WithStack(encoding.ErrInvalidValue)
						} else {
							return t.Pairs(), nil
						}
					})
				} else {
					alias := NewString(tag.alias)
					enc = encoding.EncoderFunc[reflect.Value, []Value](func(source reflect.Value) ([]Value, error) {
						if target, err := encoder.Encode(source.Field(i).Interface()); err != nil {
							return nil, err
						} else {
							return []Value{alias, target}, nil
						}
					})
				}

				encoders = append(encoders, enc)
			}

			return encoding.EncoderFunc[reflect.Value, *Map](func(source reflect.Value) (*Map, error) {
				pairs := make([]Value, 0, source.NumField()*2)
				for _, enc := range encoders {
					if targets, err := enc.Encode(source); err != nil {
						return nil, err
					} else {
						pairs = append(pairs, targets...)
					}
				}
				return NewMap(pairs...), nil
			}), nil
		}

		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	}

	return encoding.EncoderFunc[any, Value](func(source any) (Value, error) {
		s := reflect.ValueOf(source)
		if !s.IsValid() {
			return nil, errors.WithStack(encoding.ErrUnsupportedValue)
		}

		if enc, ok := typeToEncoder.Load(s.Type()); ok {
			return enc.(encoding.Encoder[reflect.Value, *Map]).Encode(s)
		}

		enc, err := compile(s.Type())
		if err != nil {
			return nil, err
		}

		typeToEncoder.Store(s.Type(), enc)
		return enc.Encode(s)
	})
}

func newMapDecoder(decoder encoding.Decoder[Value, any]) encoding.Decoder[Value, any] {
	typeToDecoder := sync.Map{} // map[reflect.Type]encoding.Decoder[*Map, reflect.Value]{}

	compile := func(typ reflect.Type) (encoding.Decoder[*Map, reflect.Value], error) {
		if typ.Kind() == reflect.Pointer {
			switch typ.Elem().Kind() {
			case reflect.Map:
				keyType := typ.Elem().Key()
				valueType := typ.Elem().Elem()

				return encoding.DecoderFunc[*Map, reflect.Value](func(source *Map, target reflect.Value) error {
					if target.Elem().IsNil() {
						target.Elem().Set(reflect.MakeMapWithSize(target.Type().Elem(), source.Len()))
					}

					for _, key := range source.Keys() {
						value, _ := source.Get(key)

						k := reflect.New(keyType)
						v := reflect.New(valueType)

						if err := decoder.Decode(key, k.Interface()); err != nil {
							return err
						} else if err := decoder.Decode(value, v.Interface()); err != nil {
							return err
						} else {
							target.Elem().SetMapIndex(k.Elem(), v.Elem())
						}
					}
					return nil
				}), nil
			case reflect.Struct:
				var decoders []encoding.Decoder[*Map, unsafe.Pointer]
				for i := 0; i < typ.Elem().NumField(); i++ {
					field := typ.Elem().Field(i)
					offset := field.Offset
					tag := getMapTag(field)

					if !field.IsExported() || tag.ignore {
						continue
					}

					var dec encoding.Decoder[*Map, unsafe.Pointer]
					if tag.inline {
						dec = encoding.DecoderFunc[*Map, unsafe.Pointer](func(source *Map, target unsafe.Pointer) error {
							return decoder.Decode(source, reflect.NewAt(field.Type, unsafe.Pointer(uintptr(target)+offset)).Interface())
						})
					} else {
						dec = encoding.DecoderFunc[*Map, unsafe.Pointer](func(source *Map, target unsafe.Pointer) error {
							value, ok := source.Get(NewString(tag.alias))
							if !ok {
								if !tag.omitempty {
									return errors.WithMessage(encoding.ErrInvalidValue, fmt.Sprintf("key(%v) is zero value", field.Name))
								}
								return nil
							}
							return decoder.Decode(value, reflect.NewAt(field.Type, unsafe.Pointer(uintptr(target)+offset)).Interface())
						})
					}

					decoders = append(decoders, dec)
				}

				return encoding.DecoderFunc[*Map, reflect.Value](func(source *Map, target reflect.Value) error {
					p := target.UnsafePointer()
					for _, dec := range decoders {
						if err := dec.Decode(source, p); err != nil {
							return err
						}
					}
					return nil
				}), nil
			default:
				if typ.Elem() == typeAny {
					return encoding.DecoderFunc[*Map, reflect.Value](func(source *Map, target reflect.Value) error {
						target.Elem().Set(reflect.ValueOf(source.Interface()))
						return nil
					}), nil
				}
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	}

	return encoding.DecoderFunc[Value, any](func(source Value, target any) error {
		if s, ok := source.(*Map); ok {
			t := reflect.ValueOf(target)

			if dec, ok := typeToDecoder.Load(t.Type()); ok {
				return dec.(encoding.Decoder[*Map, reflect.Value]).Decode(s, t)
			}

			dec, err := compile(t.Type())
			if err != nil {
				return err
			}

			typeToDecoder.Store(t.Type(), dec)
			return dec.Decode(s, t)
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func getMapTag(f reflect.StructField) mapTag {
	key := strcase.ToSnake(f.Name)
	rawTag := f.Tag.Get(tagMap)

	if rawTag != "" {
		if rawTag == "-" {
			return mapTag{ignore: true}
		}

		if index := strings.Index(rawTag, ","); index != -1 {
			tag := mapTag{}
			tag.alias = key
			if rawTag[:index] != "" {
				tag.alias = rawTag[:index]
			}

			if rawTag[index+1:] == "omitempty" {
				tag.omitempty = true
			} else if rawTag[index+1:] == "inline" {
				tag.alias = ""
				tag.inline = true
			}
			return tag
		} else {
			return mapTag{alias: rawTag}
		}
	}

	return mapTag{alias: key}
}
