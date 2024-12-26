package types

import (
	"encoding/binary"
	"hash/fnv"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// Map represents a key-value map with support for immutability and mutability.
type Map interface {
	Value

	// Get retrieves the value associated with the given key.
	Get(key Value) Value
	// Set adds or updates a key-value pair in the map.
	Set(key Value, val Value) Map
	// Delete removes a key-value pair from the map by key.
	Delete(key Value) Map
	// Clear removes all key-value pairs from the map.
	Clear() Map

	// Keys returns all keys in the map.
	Keys() []Value
	// Values returns all values in the map.
	Values() []Value
	// Pairs returns all key-value pairs in the map.
	Pairs() []Value
	// Len returns the number of key-value pairs in the map.
	Len() int

	// Range provides an iterator function for traversing key-value pairs.
	Range() func(func(key, value Value) bool)

	// Mutable returns a mutable version of the map.
	Mutable() Map
	// Immutable returns an immutable version of the map.
	Immutable() Map

	// Map converts the map into a native Go map.
	Map() map[any]any
}

type immutableMap struct {
	value map[uint64][][2]Value
	hash  uint64
	mu    sync.Mutex
}

type mutableMap struct {
	value map[uint64][][2]Value
}

type mapMeta struct {
	alias     string
	ignore    bool
	omitempty bool
	inline    bool
}

const tagMap = "map"

var _ Map = (*immutableMap)(nil)
var _ Map = (*mutableMap)(nil)

// NewMap creates a new Map with key-value pairs.
func NewMap(pairs ...Value) Map {
	value := make(map[uint64][][2]Value, len(pairs)/2)
	for i := 0; i < len(pairs)/2; i++ {
		k, v := pairs[i*2], pairs[i*2+1]

		hash := HashOf(k)
		exists := false
		if elements, ok := value[hash]; ok {
			for _, pair := range elements {
				if Equal(pair[0], k) {
					pair[1] = v
					exists = true
					break
				}
			}
		}
		if !exists {
			value[hash] = append(value[hash], [2]Value{k, v})
		}
	}

	for _, elements := range value {
		sort.Slice(elements, func(i, j int) bool {
			return Compare(elements[i][0], elements[j][0]) < 0
		})
	}

	return &immutableMap{value: value}
}

// Get retrieves the value associated with the given key.
func (m *immutableMap) Get(key Value) Value {
	if elements, ok := m.value[HashOf(key)]; ok {
		for _, pair := range elements {
			if Equal(pair[0], key) {
				return pair[1]
			}
		}
	}
	return nil
}

// Set adds or updates a key-value pair in the map.
func (m *immutableMap) Set(key, val Value) Map {
	return m.Mutable().Set(key, val).Immutable()
}

// Delete removes a key-value pair from the map by key.
func (m *immutableMap) Delete(key Value) Map {
	return m.Mutable().Delete(key).Immutable()
}

// Clear all key-value pairs in the mutable map.
func (m *immutableMap) Clear() Map {
	return &immutableMap{value: make(map[uint64][][2]Value)}
}

// Keys returns all keys in the map.
func (m *immutableMap) Keys() []Value {
	keys := make([]Value, 0, len(m.value))
	for _, elements := range m.value {
		for _, pair := range elements {
			keys = append(keys, pair[0])
		}
	}
	return keys
}

// Values returns all values in the map.
func (m *immutableMap) Values() []Value {
	values := make([]Value, 0, len(m.value))
	for _, elements := range m.value {
		for _, pair := range elements {
			values = append(values, pair[1])
		}
	}
	return values
}

// Pairs returns all key-value pairs in the map.
func (m *immutableMap) Pairs() []Value {
	pairs := make([]Value, 0, len(m.value)*2)
	for _, elements := range m.value {
		for _, pair := range elements {
			pairs = append(pairs, pair[0], pair[1])
		}
	}
	return pairs
}

// Len returns the number of key-value pairs in the map.
func (m *immutableMap) Len() int {
	length := 0
	for _, elements := range m.value {
		length += len(elements)
	}
	return length
}

// Range provides an iterator function for traversing key-value pairs.
func (m *immutableMap) Range() func(func(key, value Value) bool) {
	return func(yield func(key Value, value Value) bool) {
		keys := make([]uint64, 0, len(m.value))
		for key := range m.value {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		for _, key := range keys {
			for _, pair := range m.value[key] {
				k, v := pair[0], pair[1]
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// Immutable returns the immutable version of the map.
func (m *immutableMap) Immutable() Map {
	return m
}

// Mutable returns a mutable version of the map.
func (m *immutableMap) Mutable() Map {
	value := make(map[uint64][][2]Value, len(m.value))
	for hash, elements := range m.value {
		value[hash] = elements
	}
	return &mutableMap{value: value}
}

// Map converts the map into a native Go map.
func (m *immutableMap) Map() map[any]any {
	if len(m.value) == 0 {
		return nil
	}

	values := make(map[any]any, len(m.value))
	for _, elements := range m.value {
		for _, pair := range elements {
			k, v := pair[0], pair[1]
			values[InterfaceOf(k)] = InterfaceOf(v)
		}
	}
	return values
}

// Kind returns the kind of the map.
func (m *immutableMap) Kind() Kind {
	return KindMap
}

// Hash calculates the hash code for the map.
func (m *immutableMap) Hash() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.hash == 0 {
		keys := make([]uint64, 0, len(m.value))
		for key := range m.value {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		h := fnv.New64a()
		var buf [8]byte
		for _, key := range keys {
			for _, pair := range m.value[key] {
				k, v := pair[0], pair[1]

				binary.BigEndian.PutUint64(buf[:], HashOf(k))
				_, _ = h.Write(buf[:])

				binary.BigEndian.PutUint64(buf[:], HashOf(v))
				_, _ = h.Write(buf[:])
			}
		}
		m.hash = h.Sum64()
	}
	return m.hash
}

// Interface converts the Map to an any.
func (m *immutableMap) Interface() any {
	if len(m.value) == 0 {
		return nil
	}

	keys := make([]any, 0, len(m.value))
	values := make([]any, 0, len(m.value))

	hashable := true
	for _, elements := range m.value {
		for _, pair := range elements {
			k, v := pair[0], pair[1]

			keys = append(keys, InterfaceOf(k))
			values = append(values, InterfaceOf(v))

			kind := KindOf(k)
			hashable = hashable && kind != KindBinary && kind != KindMap && kind != KindSlice
		}
	}

	if !hashable {
		t := make([][2]any, len(keys))
		for i, key := range keys {
			value := values[i]
			t[i] = [2]any{key, value}
		}
		return t
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

// Equal checks if two Map instances are equal.
func (m *immutableMap) Equal(other Value) bool {
	if o, ok := other.(*immutableMap); ok {
		if m.Hash() != o.Hash() {
			return false
		}

		if len(m.value) == len(o.value) {
			for hash, elements1 := range m.value {
				elements2 := o.value[hash]
				if len(elements1) != len(elements2) {
					return false
				}

				for i := 0; i < len(elements1); i++ {
					v1 := elements1[i][1]
					v2 := elements2[i][1]

					if !Equal(v1, v2) {
						return false
					}
				}
			}
			return true
		}
	}
	return false
}

// Compare checks whether another Object is equal to this Map instance.
func (m *immutableMap) Compare(other Value) int {
	if o, ok := other.(*immutableMap); ok {
		if len(m.value) != len(o.value) {
			return compare(len(m.value), len(o.value))
		}

		keys := make([]uint64, 0, len(m.value))
		for key := range m.value {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		for _, hash := range keys {
			elements1 := m.value[hash]
			elements2 := o.value[hash]
			if len(elements1) != len(elements2) {
				return compare(len(elements1), len(elements2))
			}

			for i := 0; i < len(elements1); i++ {
				v1 := elements1[i][1]
				v2 := elements2[i][1]

				if c := Compare(v1, v2); c != 0 {
					return c
				}
			}
		}
		return 0
	}
	return compare(m.Kind(), KindOf(other))
}

// Get retrieves the value associated with the given key.
func (m *mutableMap) Get(key Value) Value {
	return m.Immutable().Get(key)
}

// Set adds or updates a key-value pair in the map.
func (m *mutableMap) Set(key, val Value) Map {
	hash := HashOf(key)
	exists := false
	if elements, ok := m.value[hash]; ok {
		modify := make([][2]Value, len(elements))
		copy(modify, elements)

		for _, pair := range modify {
			if Equal(pair[0], key) {
				pair[1] = val
				exists = true
				break
			}
		}

		m.value[hash] = modify
	}
	if !exists {
		m.value[hash] = append(m.value[hash], [2]Value{key, val})
	}

	elements := m.value[hash]
	sort.Slice(elements, func(i, j int) bool {
		return Compare(elements[i][0], elements[j][0]) < 0
	})

	return m
}

// Delete removes a key-value pair from the map by key.
func (m *mutableMap) Delete(key Value) Map {
	hash := HashOf(key)
	if elements, ok := m.value[hash]; ok {
		modify := make([][2]Value, len(elements))
		copy(modify, elements)

		for i, pair := range modify {
			if Equal(pair[0], key) {
				modify = append(modify[:i], modify[i+1:]...)
				break
			}
		}

		if len(modify) > 0 {
			m.value[hash] = modify
		} else {
			delete(m.value, hash)
		}
	}

	return m
}

// Clear all key-value pairs in the mutable map.
func (m *mutableMap) Clear() Map {
	m.value = make(map[uint64][][2]Value)
	return m
}

// Keys returns all keys in the map.
func (m *mutableMap) Keys() []Value {
	return m.Immutable().Keys()
}

// Values returns all values in the map.
func (m *mutableMap) Values() []Value {
	return m.Immutable().Values()
}

// Pairs returns all key-value pairs in the map.
func (m *mutableMap) Pairs() []Value {
	return m.Immutable().Pairs()
}

// Len returns the number of key-value pairs in the map.
func (m *mutableMap) Len() int {
	return m.Immutable().Len()
}

// Range provides an iterator function for traversing key-value pairs.
func (m *mutableMap) Range() func(func(key, value Value) bool) {
	return m.Immutable().Range()
}

// Immutable returns the immutable version of the map.
func (m *mutableMap) Immutable() Map {
	return &immutableMap{value: m.value}
}

// Mutable returns a mutable version of the map.
func (m *mutableMap) Mutable() Map {
	return m
}

// Map converts the map into a native Go map.
func (m *mutableMap) Map() map[any]any {
	return m.Immutable().Map()
}

// Kind returns the kind of the map.
func (m *mutableMap) Kind() Kind {
	return KindMap
}

// Hash calculates the hash code for the map.
func (m *mutableMap) Hash() uint64 {
	return m.Immutable().Hash()
}

// Interface converts the Map to an any.
func (m *mutableMap) Interface() any {
	return m.Immutable().Interface()
}

// Equal checks if two Map instances are equal.
func (m *mutableMap) Equal(other Value) bool {
	return m.Immutable().Equal(other)
}

// Compare checks whether another Object is equal to this Map instance.
func (m *mutableMap) Compare(other Value) int {
	return m.Immutable().Compare(other)
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

			return encoding.EncodeFunc(func(source any) (Value, error) {
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
			encoders := make([]encoding.Encoder[reflect.Value, []Value], 0, typ.NumField())
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)
				meta := getMapMeta(field)
				if !field.IsExported() || meta.ignore {
					continue
				}

				child, err := encoder.Compile(field.Type)
				if err != nil {
					child = encoder
				}

				alias := NewString(meta.alias)

				var enc encoding.Encoder[reflect.Value, []Value]
				if meta.inline {
					enc = encoding.EncodeFunc(func(source reflect.Value) ([]Value, error) {
						elem := source.FieldByIndex(field.Index)
						if meta.omitempty && elem.IsZero() {
							return nil, nil
						}

						if target, err := child.Encode(elem.Interface()); err != nil {
							return nil, err
						} else if t, ok := target.(Map); !ok {
							return nil, errors.WithStack(encoding.ErrUnsupportedValue)
						} else {
							return t.Pairs(), nil
						}
					})
				} else {
					enc = encoding.EncodeFunc(func(source reflect.Value) ([]Value, error) {
						elem := source.FieldByIndex(field.Index)
						if meta.omitempty && elem.IsZero() {
							return nil, nil
						}

						if target, err := child.Encode(elem.Interface()); err != nil {
							return nil, err
						} else {
							return []Value{alias, target}, nil
						}
					})
				}

				encoders = append(encoders, enc)
			}

			return encoding.EncodeFunc(func(source any) (Value, error) {
				s := reflect.ValueOf(source)
				pairs := make([]Value, 0, len(encoders)*2)
				for _, enc := range encoders {
					if target, err := enc.Encode(s); err != nil {
						return nil, err
					} else {
						pairs = append(pairs, target...)
					}
				}
				return NewMap(pairs...), nil
			}), nil
		}

		return nil, errors.WithStack(encoding.ErrUnsupportedType)
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

				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if source == nil {
						return nil
					}

					var m Map
					if s, ok := source.(Map); ok {
						m = s
					} else {
						return errors.WithStack(encoding.ErrUnsupportedType)
					}

					t := reflect.NewAt(typ.Elem(), target).Elem()
					if t.IsNil() {
						t.Set(reflect.MakeMapWithSize(t.Type(), m.Len()))
					}

					for key, value := range m.Range() {
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

					m.Clear()
					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Struct {
				var decoders []encoding.Decoder[Map, unsafe.Pointer]
				for i := 0; i < typ.Elem().NumField(); i++ {
					field := typ.Elem().Field(i)
					meta := getMapMeta(field)

					if !field.IsExported() || meta.ignore {
						continue
					}

					child, err := decoder.Compile(reflect.PointerTo(field.Type))
					if err != nil {
						return nil, err
					}

					offset := field.Offset
					alias := NewString(meta.alias)

					var dec encoding.Decoder[Map, unsafe.Pointer]
					if meta.inline {
						dec = encoding.DecodeFunc(func(source Map, target unsafe.Pointer) error {
							return child.Decode(source, unsafe.Pointer(uintptr(target)+offset))
						})
					} else {
						dec = encoding.DecodeFunc(func(source Map, target unsafe.Pointer) error {
							value := source.Get(alias)
							if value == nil {
								return nil
							}
							source.Delete(alias)
							return child.Decode(value, unsafe.Pointer(uintptr(target)+offset))
						})
					}

					decoders = append(decoders, dec)
				}

				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if source == nil {
						return nil
					}

					var m Map
					if s, ok := source.(Map); ok {
						m = s.Mutable()
					} else {
						return errors.WithStack(encoding.ErrUnsupportedType)
					}

					for _, dec := range decoders {
						if err := dec.Decode(m, target); err != nil {
							return err
						}
					}
					return nil
				}), nil
			} else if typ.Elem().Kind() == reflect.Interface {
				return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
					if s, ok := source.(Map); ok {
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

func getMapMeta(f reflect.StructField) mapMeta {
	key := strcase.ToSnake(f.Name)
	tag := f.Tag.Get(tagMap)

	if tag != "" {
		if tag == "-" {
			return mapMeta{ignore: true}
		}

		if index := strings.Index(tag, ","); index != -1 {
			meta := mapMeta{}
			meta.alias = key
			if tag[:index] != "" {
				meta.alias = tag[:index]
			}

			if tag[index+1:] == "omitempty" {
				meta.omitempty = true
			} else if tag[index+1:] == "inline" {
				meta.alias = ""
				meta.inline = true
			}
			return meta
		} else {
			return mapMeta{alias: tag}
		}
	}
	return mapMeta{alias: key}
}
