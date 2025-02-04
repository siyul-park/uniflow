package types

import (
	"encoding/binary"
	"encoding/json"
	"hash/fnv"
	"reflect"
	"slices"
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

	// Has checks if the map contains the specified key.
	Has(key Value) bool
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
var _ json.Marshaler = (*immutableMap)(nil)
var _ json.Marshaler = (*mutableMap)(nil)
var _ json.Unmarshaler = (*immutableMap)(nil)
var _ json.Unmarshaler = (*mutableMap)(nil)

// NewMap creates a new Map with key-value pairs.
func NewMap(pairs ...Value) Map {
	m := NewMapWithSize(len(pairs) / 2)
	for i := 0; i < len(pairs)/2; i++ {
		k, v := pairs[i*2], pairs[i*2+1]
		m.Set(k, v)
	}
	return m.Immutable()
}

// NewMapWithSize creates a new Map with the specified size.
func NewMapWithSize(size int) Map {
	return &mutableMap{value: make(map[uint64][][2]Value, size)}
}

// Has checks if the map contains the specified key.
func (m *immutableMap) Has(key Value) bool {
	if bucket, ok := m.value[HashOf(key)]; ok {
		low, high := 0, len(bucket)-1
		for low <= high {
			mid := low + (high-low)/2
			diff := Compare(bucket[mid][0], key)
			if diff == 0 {
				return true
			} else if diff < 0 {
				low = mid + 1
			} else {
				high = mid - 1
			}
		}
	}
	return false
}

// Get retrieves the value associated with the given key.
func (m *immutableMap) Get(key Value) Value {
	if bucket, ok := m.value[HashOf(key)]; ok {
		low, high := 0, len(bucket)-1
		for low <= high {
			mid := low + (high-low)/2
			diff := Compare(bucket[mid][0], key)
			if diff == 0 {
				return bucket[mid][1]
			} else if diff < 0 {
				low = mid + 1
			} else {
				high = mid - 1
			}
		}
	}
	return nil
}

// Set adds or updates a key-value pair in the map.
func (m *immutableMap) Set(key, val Value) Map {
	if m.Has(key) && Equal(m.Get(key), val) {
		return m
	}
	return m.mutable().Set(key, val).Immutable()
}

// Delete removes a key-value pair from the map by key.
func (m *immutableMap) Delete(key Value) Map {
	if !m.Has(key) {
		return m
	}
	return m.mutable().Delete(key).Immutable()
}

// Clear all key-value pairs in the mutable map.
func (m *immutableMap) Clear() Map {
	return &immutableMap{value: make(map[uint64][][2]Value)}
}

// Keys returns all keys in the map.
func (m *immutableMap) Keys() []Value {
	keys := make([]Value, 0, len(m.value))
	for _, bucket := range m.value {
		for _, pair := range bucket {
			keys = append(keys, pair[0])
		}
	}
	return keys
}

// Values returns all values in the map.
func (m *immutableMap) Values() []Value {
	values := make([]Value, 0, len(m.value))
	for _, bucket := range m.value {
		for _, pair := range bucket {
			values = append(values, pair[1])
		}
	}
	return values
}

// Pairs returns all key-value pairs in the map.
func (m *immutableMap) Pairs() []Value {
	pairs := make([]Value, 0, len(m.value)*2)
	for _, bucket := range m.value {
		for _, pair := range bucket {
			pairs = append(pairs, pair[0], pair[1])
		}
	}
	return pairs
}

// Len returns the number of key-value pairs in the map.
func (m *immutableMap) Len() int {
	length := 0
	for _, bucket := range m.value {
		length += len(bucket)
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
		slices.Sort(keys)

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
	return m.mutable()
}

// Map converts the map into a native Go map.
func (m *immutableMap) Map() map[any]any {
	if len(m.value) == 0 {
		return nil
	}

	values := make(map[any]any, len(m.value))
	for _, bucket := range m.value {
		for _, pair := range bucket {
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
		slices.Sort(keys)

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

	var keyType reflect.Type
	var valueType reflect.Type
	for _, bucket := range m.value {
		for _, pair := range bucket {
			keyType = unionType(keyType, TypeOf(KindOf(pair[0])))
			valueType = unionType(valueType, TypeOf(KindOf(pair[1])))
		}
	}
	if keyType == nil {
		keyType = types[KindUnknown]
	}
	if valueType == nil {
		valueType = types[KindUnknown]
	}

	if keyType.Kind() == reflect.Interface || keyType.Kind() == reflect.Map || keyType.Kind() == reflect.Slice {
		t := make([][2]any, 0, len(m.value))
		for _, bucket := range m.value {
			for _, pair := range bucket {
				t = append(t, [2]any{InterfaceOf(pair[0]), InterfaceOf(pair[1])})
			}
		}
		return t
	}

	t := reflect.MakeMapWithSize(reflect.MapOf(keyType, valueType), len(m.value))
	for _, bucket := range m.value {
		for _, pair := range bucket {
			t.SetMapIndex(reflect.ValueOf(InterfaceOf(pair[0])), reflect.ValueOf(InterfaceOf(pair[1])))
		}
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
		slices.Sort(keys)

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

// MarshalJSON converts the map into a JSON byte array.
func (m *immutableMap) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, 1024)
	buf = append(buf, '{')
	for k, v := range m.Range() {
		if len(buf) > 1 {
			buf = append(buf, ',')
		}

		key, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf = append(buf, key...)
		buf = append(buf, ':')

		value, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		buf = append(buf, value...)
	}
	buf = append(buf, '}')
	return buf, nil
}

// UnmarshalJSON converts the JSON byte array into a map.
func (m *immutableMap) UnmarshalJSON(bytes []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mutable := &mutableMap{value: make(map[uint64][][2]Value)}
	if err := mutable.UnmarshalJSON(bytes); err != nil {
		return err
	}
	m.value = mutable.value
	m.hash = 0
	return nil
}

func (m *immutableMap) mutable() *mutableMap {
	value := make(map[uint64][][2]Value, len(m.value))
	for hash, bucket := range m.value {
		value[hash] = bucket
	}
	return &mutableMap{value: value}
}

// Has checks if the map contains the specified key.
func (m *mutableMap) Has(key Value) bool {
	return m.Immutable().Has(key)
}

// Get retrieves the value associated with the given key.
func (m *mutableMap) Get(key Value) Value {
	return m.Immutable().Get(key)
}

// Set adds or updates a key-value pair in the map.
func (m *mutableMap) Set(key, val Value) Map {
	hash := HashOf(key)
	bucket := m.value[hash]

	diff := -1
	low, high := 0, len(bucket)-1
	for low <= high {
		mid := low + (high-low)/2
		if diff = Compare(bucket[mid][0], key); diff == 0 {
			modify := make([][2]Value, len(bucket))
			copy(modify, bucket)
			bucket[mid][1] = val

			m.value[hash] = modify
			break
		} else if diff < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	if diff != 0 {
		modify := make([][2]Value, len(bucket)+1)
		copy(modify[:low], bucket[:low])
		copy(modify[low+1:], bucket[low:])
		modify[low] = [2]Value{key, val}

		m.value[hash] = modify
	}

	return m
}

// Delete removes a key-value pair from the map by key.
func (m *mutableMap) Delete(key Value) Map {
	hash := HashOf(key)
	if bucket, ok := m.value[hash]; ok {
		low, high := 0, len(bucket)-1
		for low <= high {
			mid := low + (high-low)/2
			diff := Compare(bucket[mid][0], key)
			if diff == 0 {
				modify := make([][2]Value, len(bucket)-1)
				copy(modify[:mid], bucket[:mid])
				copy(modify[mid:], bucket[mid+1:])

				if len(modify) > 0 {
					m.value[hash] = modify
				} else {
					delete(m.value, hash)
				}
				break
			} else if diff < 0 {
				low = mid + 1
			} else {
				high = mid - 1
			}
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
	return m.immutable().Keys()
}

// Values returns all values in the map.
func (m *mutableMap) Values() []Value {
	return m.immutable().Values()
}

// Pairs returns all key-value pairs in the map.
func (m *mutableMap) Pairs() []Value {
	return m.immutable().Pairs()
}

// Len returns the number of key-value pairs in the map.
func (m *mutableMap) Len() int {
	return m.immutable().Len()
}

// Range provides an iterator function for traversing key-value pairs.
func (m *mutableMap) Range() func(func(key, value Value) bool) {
	return m.immutable().Range()
}

// Immutable returns the immutable version of the map.
func (m *mutableMap) Immutable() Map {
	return m.immutable()
}

// Mutable returns a mutable version of the map.
func (m *mutableMap) Mutable() Map {
	return m
}

// Map converts the map into a native Go map.
func (m *mutableMap) Map() map[any]any {
	return m.immutable().Map()
}

// Kind returns the kind of the map.
func (m *mutableMap) Kind() Kind {
	return KindMap
}

// Hash calculates the hash code for the map.
func (m *mutableMap) Hash() uint64 {
	return m.immutable().Hash()
}

// Interface converts the Map to an any.
func (m *mutableMap) Interface() any {
	return m.immutable().Interface()
}

// Equal checks if two Map instances are equal.
func (m *mutableMap) Equal(other Value) bool {
	return m.immutable().Equal(other)
}

// Compare checks whether another Object is equal to this Map instance.
func (m *mutableMap) Compare(other Value) int {
	return m.immutable().Compare(other)
}

// MarshalJSON converts the map into a JSON byte array.
func (m *mutableMap) MarshalJSON() ([]byte, error) {
	return m.immutable().MarshalJSON()
}

// UnmarshalJSON converts the JSON byte array into a map.
func (m *mutableMap) UnmarshalJSON(bytes []byte) error {
	var value map[string]any
	if err := json.Unmarshal(bytes, &value); err != nil {
		return err
	}

	for k, v := range value {
		key := NewString(k)
		val, err := Marshal(v)
		if err != nil {
			return err
		}
		m.Set(key, val)
	}
	return nil
}

func (m *mutableMap) immutable() *immutableMap {
	return &immutableMap{value: m.value}
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
				m := NewMapWithSize(s.Len())
				for _, k := range s.MapKeys() {
					v := s.MapIndex(k)
					if key, err := keyEncoder.Encode(k.Interface()); err != nil {
						return nil, err
					} else if val, err := valueEncoder.Encode(v.Interface()); err != nil {
						return nil, err
					} else {
						m.Set(key, val)
					}
				}
				return m.Immutable(), nil
			}), nil
		} else if typ != nil && typ.Kind() == reflect.Struct {
			decoders := make([]encoding.Decoder[reflect.Value, Map], 0, typ.NumField())
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

				var dec encoding.Decoder[reflect.Value, Map]
				if meta.inline {
					dec = encoding.DecodeFunc(func(source reflect.Value, m Map) error {
						elem := source.FieldByIndex(field.Index)
						if meta.omitempty && elem.IsZero() {
							return nil
						}

						if target, err := child.Encode(elem.Interface()); err != nil {
							return err
						} else if t, ok := target.(Map); !ok {
							return errors.WithStack(encoding.ErrUnsupportedValue)
						} else {
							for k, v := range t.Range() {
								m.Set(k, v)
							}
							return nil
						}
					})
				} else {
					dec = encoding.DecodeFunc(func(source reflect.Value, m Map) error {
						elem := source.FieldByIndex(field.Index)
						if meta.omitempty && elem.IsZero() {
							return nil
						}

						if target, err := child.Encode(elem.Interface()); err != nil {
							return err
						} else {
							m.Set(alias, target)
							return nil
						}
					})
				}

				decoders = append(decoders, dec)
			}

			return encoding.EncodeFunc(func(source any) (Value, error) {
				s := reflect.ValueOf(source)
				m := NewMapWithSize(len(decoders) * 2)
				for _, dec := range decoders {
					if err := dec.Decode(s, m); err != nil {
						return nil, err
					}
				}
				return m.Immutable(), nil
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
			} else if typ.Elem() == types[KindUnknown] {
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
