package primitive

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/benbjohnson/immutable"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
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

	// If the types are different, compare based on type kind.
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
			values = append(values, v.Interface())
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
	return encoding.EncoderFunc[any, Value](func(source any) (Value, error) {
		if s := reflect.ValueOf(source); s.Kind() == reflect.Map {
			pairs := make([]Value, len(s.MapKeys())*2)
			for i, k := range s.MapKeys() {
				if k, err := encoder.Encode(k.Interface()); err != nil {
					return nil, errors.WithMessage(err, fmt.Sprintf("key(%v) can't encode", k.Interface()))
				} else {
					pairs[i*2] = k
				}
				if v, err := encoder.Encode(s.MapIndex(k).Interface()); err != nil {
					return nil, errors.WithMessage(err, fmt.Sprintf("value(%v) can't encode", s.MapIndex(k).Interface()))
				} else {
					pairs[i*2+1] = v
				}
			}
			return NewMap(pairs...), nil
		} else if s := reflect.ValueOf(source); s.Kind() == reflect.Struct {
			pairs := make([]Value, 0, s.NumField()*2)
			for i := 0; i < s.NumField(); i++ {
				field := s.Type().Field(i)
				if !field.IsExported() {
					continue
				}

				v := s.FieldByName(field.Name)
				tag := getMapTag(s.Type(), field)

				if tag.ignore || (tag.omitempty && v.IsZero()) {
					continue
				}

				if v, err := encoder.Encode(v.Interface()); err != nil {
					return nil, errors.WithMessage(err, fmt.Sprintf("field(%s) can't encode", field.Name))
				} else {
					if tag.inline {
						if v, ok := v.(*Map); ok {
							for _, k := range v.Keys() {
								pairs = append(pairs, k)
								pairs = append(pairs, v.GetOr(k, nil))
							}
						} else {
							return nil, errors.WithStack(encoding.ErrUnsupportedValue)
						}
					} else {
						pairs = append(pairs, NewString(tag.alias))
						pairs = append(pairs, v)
					}
				}
			}
			return NewMap(pairs...), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newMapDecoder(decoder encoding.Decoder[Value, any]) encoding.Decoder[Value, any] {
	return encoding.DecoderFunc[Value, any](func(source Value, target any) error {
		if s, ok := source.(*Map); ok {
			if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				switch t.Elem().Kind() {
				case reflect.Map:
					if t.Elem().IsNil() {
						t.Elem().Set(reflect.MakeMapWithSize(t.Type().Elem(), s.Len()))
					}

					keyType := t.Elem().Type().Key()
					valueType := t.Elem().Type().Elem()

					for _, key := range s.Keys() {
						value, _ := s.Get(key)

						k := reflect.New(keyType)
						v := reflect.New(valueType)

						if err := decoder.Decode(key, k.Interface()); err != nil {
							return errors.WithMessage(err, fmt.Sprintf("key(%v) cannot be decoded", key.Interface()))
						} else if err := decoder.Decode(value, v.Interface()); err != nil {
							return errors.WithMessage(err, fmt.Sprintf("value(%v) corresponding to the key(%v) cannot be decoded", value.Interface(), key.Interface()))
						}

						t.Elem().SetMapIndex(k.Elem(), v.Elem())
					}
					return nil
				case reflect.Struct:
					for i := 0; i < t.Elem().NumField(); i++ {
						field := t.Elem().Type().Field(i)
						if !field.IsExported() {
							continue
						}

						v := t.Elem().FieldByName(field.Name)
						tag := getMapTag(t.Type().Elem(), field)

						if tag.ignore {
							continue
						} else if tag.inline {
							if err := decoder.Decode(s, v.Addr().Interface()); err != nil {
								return err
							} else {
								continue
							}
						}

						value, ok := s.Get(NewString(tag.alias))
						if !ok {
							if !tag.omitempty {
								return errors.WithMessage(encoding.ErrUnsupportedValue, fmt.Sprintf("key(%v) is zero value", field.Name))
							}
						} else if err := decoder.Decode(value, v.Addr().Interface()); err != nil {
							return errors.WithMessage(err, fmt.Sprintf("value(%v) corresponding to the key(%v) cannot be decoded", value.Interface(), field.Name))
						}
					}
				default:
					if t.Type() == typeAny {
						t.Elem().Set(reflect.ValueOf(s.Interface()))
					} else {
						return errors.WithStack(encoding.ErrUnsupportedValue)
					}
				}
				return nil
			}
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func getMapTag(t reflect.Type, f reflect.StructField) mapTag {
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
