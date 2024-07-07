package types

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
type Map = *_map

type _map struct {
	value *immutable.SortedMap[Value, Value]
}

type mapTag struct {
	alias     string
	ignore    bool
	omitempty bool
	inline    bool
}

type comparer struct{}

const tagMap = "map"

var _ Value = (Map)(nil)
var _ immutable.Comparer[Value] = &comparer{}

// NewMap creates a new Map with key-value pairs.
func NewMap(pairs ...Value) Map {
	b := immutable.NewSortedMapBuilder[Value, Value](&comparer{})
	for i := 0; i < len(pairs)/2; i++ {
		k, v := pairs[i*2], pairs[i*2+1]
		b.Set(k, v)
	}
	return &_map{value: b.Map()}
}

// Get retrieves the value for a given key.
func (m Map) Get(key Value) (Value, bool) {
	return m.value.Get(key)
}

// GetOr returns the value for a given key or a default value if the key is not found.
func (m Map) GetOr(key, value Value) Value {
	if v, ok := m.value.Get(key); ok {
		return v
	}
	return value
}

// Set adds or updates a key-value pair in the map.
func (m Map) Set(key, value Value) Map {
	return &_map{value: m.value.Set(key, value)}
}

// Delete removes a key and its corresponding value from the map.
func (m Map) Delete(key Value) Map {
	return &_map{value: m.value.Delete(key)}
}

// Keys returns all keys in the map.
func (m Map) Keys() []Value {
	keys := make([]Value, 0, m.value.Len())
	for itr := m.value.Iterator(); !itr.Done(); {
		k, _, _ := itr.Next()
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values in the map.
func (m Map) Values() []Value {
	values := make([]Value, 0, m.value.Len())
	for itr := m.value.Iterator(); !itr.Done(); {
		_, v, _ := itr.Next()
		values = append(values, v)
	}
	return values
}

// Pairs returns all keys and values in the map.
func (m Map) Pairs() []Value {
	pairs := make([]Value, 0, m.value.Len()*2)
	for itr := m.value.Iterator(); !itr.Done(); {
		k, v, _ := itr.Next()
		pairs = append(pairs, k)
		pairs = append(pairs, v)
	}
	return pairs
}

// Len returns the number of key-value pairs in the map.
func (m Map) Len() int {
	return m.value.Len()
}

// Map converts the Map to a raw Go map.
func (m Map) Map() map[any]any {
	values := make(map[any]any, m.value.Len())
	for itr := m.value.Iterator(); !itr.Done(); {
		k, v, _ := itr.Next()
		values[InterfaceOf(k)] = InterfaceOf(v)
	}
	return values
}

// Kind returns the kind of the Map.
func (m Map) Kind() Kind {
	return KindMap
}

// Hash calculates and returns the hash code.
func (m Map) Hash() uint64 {
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
func (m Map) Interface() any {
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
func (m Map) Equal(other Value) bool {
	if o, ok := other.(Map); ok {
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
func (m Map) Compare(other Value) int {
	if o, ok := other.(Map); ok {
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

func (*comparer) Compare(x, y Value) int {
	return Compare(x, y)
}

func newMapEncoder(encoder *encoding.EncodeAssembler[any, Value]) encoding.EncodeCompiler[any, Value] {
	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ != nil && typ.Kind() == reflect.Map {
			keyType := typ.Key()
			valueType := typ.Elem()

			keyEncoder, _ := encoder.Compile(keyType)
			if keyEncoder == nil {
				keyEncoder = encoder
			}
			valueEncoder, _ := encoder.Compile(valueType)
			if valueEncoder == nil {
				valueEncoder = encoder
			}

			return encoding.EncodeFunc[any, Value](func(source any) (Value, error) {
				s := reflect.ValueOf(source)
				pairs := make([]Value, 0, s.Len()*2)
				for _, k := range s.MapKeys() {
					v := s.MapIndex(k)

					if key, err := keyEncoder.Encode(k.Interface()); err != nil {
						return nil, err
					} else {
						pairs = append(pairs, key)
					}

					if val, err := valueEncoder.Encode(v.Interface()); err != nil {
						return nil, err
					} else {
						pairs = append(pairs, val)
					}
				}
				return NewMap(pairs...), nil
			}), nil
		} else if typ != nil && typ.Kind() == reflect.Struct {
			valueEncoders := make([]encoding.Encoder[any, Value], 0, typ.NumField())
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)

				valueEncoder, _ := encoder.Compile(field.Type)
				if valueEncoder == nil {
					valueEncoder = encoder
				}
				valueEncoders = append(valueEncoders, valueEncoder)
			}

			return encoding.EncodeFunc[any, Value](func(source any) (Value, error) {
				s := reflect.ValueOf(source)

				pairs := make([]Value, 0, typ.NumField()*2)
				for i := 0; i < typ.NumField(); i++ {
					field := typ.Field(i)
					tag := getMapTag(field)

					if !field.IsExported() || tag.ignore {
						continue
					}

					elem := s.Field(i)
					valueEncoder := valueEncoders[i]

					if tag.inline {
						if target, err := valueEncoder.Encode(elem.Interface()); err != nil {
							return nil, err
						} else if t, ok := target.(Map); !ok {
							return nil, errors.WithStack(encoding.ErrInvalidArgument)
						} else {
							pairs = append(pairs, t.Pairs()...)
						}
					} else {
						if tag.omitempty && elem.IsZero() {
							continue
						}

						if target, err := valueEncoder.Encode(elem.Interface()); err != nil {
							return nil, err
						} else {
							pairs = append(pairs, NewString(tag.alias), target)
						}
					}
				}

				return NewMap(pairs...), nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}

func newMapDecoder(decoder *encoding.DecodeAssembler[Value, any]) encoding.DecodeCompiler[Value] {
	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer {
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

				return encoding.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if source == nil {
						return nil
					}
					if s, ok := source.(Map); ok {
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
					return errors.WithStack(encoding.ErrInvalidArgument)
				}), nil
			} else if typ.Elem().Kind() == reflect.Struct {
				var decoders []encoding.Decoder[Map, unsafe.Pointer]
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

					var dec encoding.Decoder[Map, unsafe.Pointer]
					if tag.inline {
						dec = encoding.DecodeFunc[Map, unsafe.Pointer](func(source Map, target unsafe.Pointer) error {
							return child.Decode(source, unsafe.Pointer(uintptr(target)+offset))
						})
					} else {
						dec = encoding.DecodeFunc[Map, unsafe.Pointer](func(source Map, target unsafe.Pointer) error {
							value, ok := source.Get(alias)
							if !ok {
								if !tag.omitempty {
									return errors.WithMessage(encoding.ErrInvalidArgument, fmt.Sprintf("key(%v) is zero value", field.Name))
								}
								return nil
							}
							return child.Decode(value, unsafe.Pointer(uintptr(target)+offset))
						})
					}

					decoders = append(decoders, dec)
				}

				return encoding.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if source == nil {
						return nil
					}
					if s, ok := source.(Map); ok {
						for _, dec := range decoders {
							if err := dec.Decode(s, target); err != nil {
								return err
							}
						}
						return nil
					}
					return errors.WithStack(encoding.ErrInvalidArgument)
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Map); ok {
						*(*any)(target) = s.Interface()
						return nil
					}
					return errors.WithStack(encoding.ErrInvalidArgument)
				}), nil
			}
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
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
