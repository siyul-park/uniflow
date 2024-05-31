package object

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"reflect"
	"strings"
	"unsafe"

	"github.com/benbjohnson/immutable"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Map represents a map structure.
type Map struct {
	value *immutable.SortedMap[Object, Object]
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

var _ Object = (*Map)(nil)
var _ immutable.Comparer[Object] = &comparer{}

// NewMap creates a new Map with key-value pairs.
func NewMap(pairs ...Object) *Map {
	b := immutable.NewSortedMapBuilder[Object, Object](&comparer{})
	for i := 0; i < len(pairs)/2; i++ {
		k, v := pairs[i*2], pairs[i*2+1]
		b.Set(k, v)
	}
	return &Map{value: b.Map()}
}

// Get retrieves the value for a given key.
func (m *Map) Get(key Object) (Object, bool) {
	return m.value.Get(key)
}

// GetOr returns the value for a given key or a default value if the key is not found.
func (m *Map) GetOr(key, value Object) Object {
	if v, ok := m.value.Get(key); ok {
		return v
	}
	return value
}

// Set adds or updates a key-value pair in the map.
func (m *Map) Set(key, value Object) *Map {
	return &Map{value: m.value.Set(key, value)}
}

// Delete removes a key and its corresponding value from the map.
func (m *Map) Delete(key Object) *Map {
	return &Map{value: m.value.Delete(key)}
}

// Keys returns all keys in the map.
func (m *Map) Keys() []Object {
	keys := make([]Object, 0, m.value.Len())
	for itr := m.value.Iterator(); !itr.Done(); {
		k, _, _ := itr.Next()
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values in the map.
func (m *Map) Values() []Object {
	values := make([]Object, 0, m.value.Len())
	for itr := m.value.Iterator(); !itr.Done(); {
		_, v, _ := itr.Next()
		values = append(values, v)
	}
	return values
}

// Pairs returns all keys and values in the map.
func (m *Map) Pairs() []Object {
	pairs := make([]Object, 0, m.value.Len()*2)
	for itr := m.value.Iterator(); !itr.Done(); {
		k, v, _ := itr.Next()
		pairs = append(pairs, k)
		pairs = append(pairs, v)
	}
	return pairs
}

// Len returns the number of key-value pairs in the map.
func (m *Map) Len() int {
	return m.value.Len()
}

// Map converts the Map to a raw Go map.
func (m *Map) Map() map[any]any {
	values := make(map[any]any, m.value.Len())
	for itr := m.value.Iterator(); !itr.Done(); {
		k, v, _ := itr.Next()
		values[InterfaceOf(k)] = InterfaceOf(v)
	}
	return values
}

// Kind returns the kind of the Map.
func (m *Map) Kind() Kind {
	return KindMap
}

// Hash calculates and returns the hash code.
func (m *Map) Hash() uint64 {
	h := fnv.New64a()
	var buf [8]byte
	for itr := m.value.Iterator(); !itr.Done(); {
		k, v, _ := itr.Next()

		binary.BigEndian.PutUint64(buf[:], HashOf(k))
		_, _ = h.Write(buf[:])

		binary.BigEndian.PutUint64(buf[:], HashOf(v))
		_, _ = h.Write(buf[:])
	}
	return h.Sum64()
}

// Interface converts the Map to an interface{}.
func (m *Map) Interface() any {
	keys := make([]any, 0, m.value.Len())
	values := make([]any, 0, m.value.Len())

	for itr := m.value.Iterator(); !itr.Done(); {
		k, v, _ := itr.Next()
		keys = append(keys, InterfaceOf(k))
		values = append(values, InterfaceOf(v))
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

// Compare compares two maps.
func (m *Map) Equal(other Object) bool {
	if o, ok := other.(*Map); ok {
		if m.value.Len() == o.value.Len() {
			itr1 := m.value.Iterator()
			itr2 := o.value.Iterator()
			for !itr1.Done() && !itr2.Done() {
				k1, v1, _ := itr1.Next()
				k2, v2, _ := itr2.Next()

				if !Equal(k1, k2) || !Equal(v1, v2) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// Compare checks whether another Object is equal to this Map instance.
func (m *Map) Compare(other Object) int {
	if o, ok := other.(*Map); ok {
		itr1 := m.value.Iterator()
		itr2 := o.value.Iterator()
		for !itr1.Done() && !itr2.Done() {
			k1, v1, _ := itr1.Next()
			k2, v2, _ := itr2.Next()

			if c := Compare(k1, k2); c != 0 {
				return c
			}
			if c := Compare(v1, v2); c != 0 {
				return c
			}
		}
		return compare(m.value.Len(), o.value.Len())
	}
	return compare(m.Kind(), KindOf(other))
}

func (*comparer) Compare(x, y Object) int {
	return Compare(x, y)
}

func newMapEncoder(encoder *encoding.EncodeAssembler[any, Object]) encoding.EncodeCompiler[Object] {
	return encoding.EncodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[unsafe.Pointer, Object], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Map {
				keyType := reflect.PointerTo(typ.Elem().Key())
				valueType := reflect.PointerTo(typ.Elem().Elem())

				keyEncoder, _ := encoder.Compile(keyType)
				valueEncoder, _ := encoder.Compile(valueType)

				return encoding.EncodeFunc[unsafe.Pointer, Object](func(source unsafe.Pointer) (Object, error) {
					t := reflect.NewAt(typ.Elem(), source).Elem()

					pairs := make([]Object, 0, t.Len()*2)

					var err error
					for _, k := range t.MapKeys() {
						v := t.MapIndex(k)

						k = reflect.ValueOf(k.Interface())
						v = reflect.ValueOf(v.Interface())

						kPtr := reflect.New(k.Type())
						kPtr.Elem().Set(k)

						vPtr := reflect.New(v.Type())
						vPtr.Elem().Set(v)

						var key Object
						if keyEncoder != nil && k.Type() == keyType.Elem() {
							key, err = keyEncoder.Encode(kPtr.UnsafePointer())
						} else {
							key, err = encoder.Encode(kPtr.Interface())
						}
						if err != nil {
							return nil, err
						}

						var val Object
						if valueEncoder != nil && v.Type() == valueType.Elem() {
							val, err = valueEncoder.Encode(vPtr.UnsafePointer())
						} else {
							val, err = encoder.Encode(vPtr.Interface())
						}
						if err != nil {
							return nil, err
						}

						pairs = append(pairs, key)
						pairs = append(pairs, val)
					}

					return NewMap(pairs...), nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Struct {
				var encoders []encoding.Encoder[unsafe.Pointer, []Object]

				for i := 0; i < typ.Elem().NumField(); i++ {
					field := typ.Elem().Field(i)
					tag := getMapTag(field)

					if !field.IsExported() || tag.ignore {
						continue
					}

					child, err := encoder.Compile(reflect.PointerTo(field.Type))
					if err != nil {
						return nil, err
					}

					offset := field.Offset
					alias := NewString(tag.alias)

					var enc encoding.Encoder[unsafe.Pointer, []Object]
					if tag.inline {
						enc = encoding.EncodeFunc[unsafe.Pointer, []Object](func(source unsafe.Pointer) ([]Object, error) {
							if target, err := child.Encode(unsafe.Pointer(uintptr(source) + offset)); err != nil {
								return nil, err
							} else if t, ok := target.(*Map); !ok {
								return nil, errors.WithStack(encoding.ErrInvalidValue)
							} else {
								return t.Pairs(), nil
							}
						})
					} else {
						enc = encoding.EncodeFunc[unsafe.Pointer, []Object](func(source unsafe.Pointer) ([]Object, error) {
							t := unsafe.Pointer(uintptr(source) + offset)
							if tag.omitempty {
								if t := reflect.NewAt(field.Type, t).Elem(); t.IsZero() {
									return nil, nil
								}
							}

							if target, err := child.Encode(t); err != nil {
								return nil, err
							} else {
								return []Object{alias, target}, nil
							}
						})
					}

					encoders = append(encoders, enc)
				}

				return encoding.EncodeFunc[unsafe.Pointer, Object](func(target unsafe.Pointer) (Object, error) {
					pairs := make([]Object, 0, len(encoders)*2)
					for _, enc := range encoders {
						if target, err := enc.Encode(target); err != nil {
							return nil, err
						} else {
							pairs = append(pairs, target...)
						}
					}
					return NewMap(pairs...), nil
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newMapDecoder(decoder *encoding.DecodeAssembler[Object, any]) encoding.DecodeCompiler[Object] {
	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer {
			if typ.Elem().Kind() == reflect.Map {
				keyType := typ.Elem().Key()
				valueType := typ.Elem().Elem()

				keyDecoder, err := decoder.Compile(reflect.PointerTo(keyType))
				if err != nil {
					return nil, err
				}
				valueDecoder, err := decoder.Compile(reflect.PointerTo(valueType))
				if err != nil {
					return nil, err
				}

				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(*Map); ok {
						t := reflect.NewAt(typ.Elem(), target).Elem()
						if t.IsNil() {
							t.Set(reflect.MakeMapWithSize(t.Type(), s.Len()))
						}

						for _, key := range s.Keys() {
							value, _ := s.Get(key)

							k := reflect.New(keyType)
							v := reflect.New(valueType)

							if err := keyDecoder.Decode(key, k.UnsafePointer()); err != nil {
								return err
							} else if err := valueDecoder.Decode(value, v.UnsafePointer()); err != nil {
								return err
							} else {
								t.SetMapIndex(k.Elem(), v.Elem())
							}
						}
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}), nil
			} else if typ.Elem().Kind() == reflect.Struct {
				var decoders []encoding.Decoder[*Map, unsafe.Pointer]
				for i := 0; i < typ.Elem().NumField(); i++ {
					field := typ.Elem().Field(i)
					tag := getMapTag(field)

					if !field.IsExported() || tag.ignore {
						continue
					}

					child, err := decoder.Compile(reflect.PointerTo(field.Type))
					if err != nil {
						return nil, err
					}

					offset := field.Offset
					alias := NewString(tag.alias)

					var dec encoding.Decoder[*Map, unsafe.Pointer]
					if tag.inline {
						dec = encoding.DecodeFunc[*Map, unsafe.Pointer](func(source *Map, target unsafe.Pointer) error {
							return child.Decode(source, unsafe.Pointer(uintptr(target)+offset))
						})
					} else {
						dec = encoding.DecodeFunc[*Map, unsafe.Pointer](func(source *Map, target unsafe.Pointer) error {
							value, ok := source.Get(alias)
							if !ok {
								if !tag.omitempty {
									return errors.WithMessage(encoding.ErrInvalidValue, fmt.Sprintf("key(%v) is zero value", field.Name))
								}
								return nil
							}
							return child.Decode(value, unsafe.Pointer(uintptr(target)+offset))
						})
					}

					decoders = append(decoders, dec)
				}

				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(*Map); ok {
						for _, dec := range decoders {
							if err := dec.Decode(s, target); err != nil {
								return err
							}
						}
						return nil
					}
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
					if s, ok := source.(*Map); ok {
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
