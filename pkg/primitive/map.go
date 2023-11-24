package primitive

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"strings"
	"unsafe"

	"github.com/benbjohnson/immutable"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/internal/encoding"
)

type (
	// Map is a representation of a map.
	Map struct {
		value *immutable.SortedMap[Object, Object]
	}

	mapTag struct {
		alias     string
		ignore    bool
		omitempty bool
		inline    bool
	}

	comparer struct{}
)

const (
	tagMap = "map"
)

var _ Object = (*Map)(nil)
var _ immutable.Comparer[Object] = (*comparer)(nil)

// NewMap returns a new Map.
func NewMap(pairs ...Object) *Map {
	b := immutable.NewSortedMapBuilder[Object, Object](&comparer{})
	for i := 0; i < len(pairs)/2; i++ {
		k := pairs[i*2]
		v := pairs[i*2+1]

		b.Set(k, v)
	}
	return &Map{value: b.Map()}
}

func (o *Map) Get(key Object) (Object, bool) {
	return o.value.Get(key)
}

func (o *Map) GetOr(key, value Object) Object {
	if v, ok := o.Get(key); ok {
		return v
	}
	return value
}

func (o *Map) Set(key, value Object) *Map {
	return &Map{value: o.value.Set(key, value)}
}

func (o *Map) Delete(key Object) *Map {
	return &Map{value: o.value.Delete(key)}
}

func (o *Map) Keys() []Object {
	var keys []Object

	itr := o.value.Iterator()
	for !itr.Done() {
		k, _, _ := itr.Next()
		keys = append(keys, k)
	}
	return keys
}

func (o *Map) Len() int {
	return o.value.Len()
}

// Map returns a raw representation.
func (o *Map) Map() map[any]any {
	m := make(map[any]any, o.value.Len())

	itr := o.value.Iterator()
	for !itr.Done() {
		k, v, _ := itr.Next()

		// FIXME: check interface is can't be map key.
		if k != nil {
			if v != nil {
				m[k.Interface()] = v.Interface()
			} else {
				m[k.Interface()] = nil
			}
		}
	}

	return m
}

func (o *Map) Kind() Kind {
	return KindMap
}

func (o *Map) Equal(v Object) bool {
	if r, ok := v.(*Map); !ok {
		return false
	} else if o.Len() != r.Len() {
		return false
	} else {
		keys1 := o.Keys()
		keys2 := r.Keys()

		for i, k1 := range keys1 {
			k2 := keys2[i]
			if !Equal(k1, k2) {
				return false
			}

			v1, ok1 := o.Get(k1)
			v2, ok2 := o.Get(k2)
			if ok1 != ok2 {
				return false
			}
			if !ok1 || !ok2 {
				continue
			}
			if !Equal(v1, v2) {
				return false
			}
		}

		return true
	}
}

func (o *Map) Hash() uint32 {
	h := fnv.New32()
	h.Write([]byte{byte(KindMap), 0})

	itr := o.value.Iterator()
	for !itr.Done() {
		k, v, _ := itr.Next()

		if k != nil {
			hash := k.Hash()
			buf := *(*[unsafe.Sizeof(hash)]byte)(unsafe.Pointer(&hash))
			h.Write(buf[:])
		} else {
			h.Write([]byte{0})
		}
		if v != nil {
			hash := v.Hash()
			buf := *(*[unsafe.Sizeof(hash)]byte)(unsafe.Pointer(&hash))
			h.Write(buf[:])
		} else {
			h.Write([]byte{0})
		}
	}

	return h.Sum32()
}

func (o *Map) Interface() any {
	var keys []any
	var values []any

	itr := o.value.Iterator()
	for !itr.Done() {
		k, v, _ := itr.Next()

		// FIXME: check interface is can't be map key.
		if k != nil {
			keys = append(keys, k.Interface())
			if v != nil {
				values = append(values, v.Interface())
			} else {
				values = append(values, nil)
			}
		}
	}

	keyType := typeAny
	valueType := typeAny

	for i, key := range keys {
		typ := reflect.TypeOf(key)
		if i == 0 {
			keyType = typ
		} else if keyType != typ {
			keyType = typeAny
		}
	}
	for i, value := range values {
		typ := reflect.TypeOf(value)
		if i == 0 {
			valueType = typ
		} else if valueType != typ {
			valueType = typeAny
		}
	}

	t := reflect.MakeMapWithSize(reflect.MapOf(keyType, valueType), len(keys))
	for i, key := range keys {
		value := values[i]
		t.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}
	return t.Interface()
}

func (*comparer) Compare(a Object, b Object) int {
	if a == nil {
		return -1
	} else if b == nil {
		return 1
	} else if a.Kind() > b.Kind() {
		return 1
	} else if a.Kind() < b.Kind() {
		return -1
	}

	hashA := a.Hash()
	hashB := b.Hash()

	if hashA > hashB {
		return 1
	} else if hashA < hashB {
		return -1
	}

	if !a.Equal(b) {
		return 1
	}
	return 0
}

// NewMapEncoder is encode map or struct to Map.
func NewMapEncoder(encoder encoding.Encoder[any, Object]) encoding.Encoder[any, Object] {
	return encoding.EncoderFunc[any, Object](func(source any) (Object, error) {
		if s := reflect.ValueOf(source); s.Kind() == reflect.Map {
			pairs := make([]Object, len(s.MapKeys())*2)
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
			pairs := make([]Object, 0, s.NumField()*2)
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

// NewMapDecoder is decode Map to map or struct.
func NewMapDecoder(decoder encoding.Decoder[Object, any]) encoding.Decoder[Object, any] {
	return encoding.DecoderFunc[Object, any](func(source Object, target any) error {
		if s, ok := source.(*Map); ok {
			if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				if t.Elem().Kind() == reflect.Map {
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
				} else if t.Elem().Kind() == reflect.Struct {
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
						if !ok || reflect.ValueOf(value.Interface()).IsZero() {
							if tag.omitempty {
								continue
							} else {
								return errors.WithMessage(encoding.ErrUnsupportedValue, fmt.Sprintf("key(%v) is zero value", field.Name))
							}
						} else if err := decoder.Decode(value, v.Addr().Interface()); err != nil {
							return errors.WithMessage(err, fmt.Sprintf("value(%v) corresponding to the key(%v) cannot be decoded", value.Interface(), field.Name))
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

func getMapTag(t reflect.Type, f reflect.StructField) mapTag {
	k := strcase.ToSnake(f.Name)
	tag := f.Tag.Get(tagMap)

	if tag != "" {
		if tag == "-" {
			return mapTag{
				ignore: true,
			}
		}

		if index := strings.Index(tag, ","); index != -1 {
			mtag := mapTag{}
			mtag.alias = k
			if tag[:index] != "" {
				mtag.alias = tag[:index]
			}

			if tag[index+1:] == "omitempty" {
				mtag.omitempty = true
			} else if tag[index+1:] == "inline" {
				mtag.alias = ""
				mtag.inline = true
			}
			return mtag
		} else {
			return mapTag{
				alias: tag,
			}
		}
	}

	return mapTag{
		alias: k,
	}
}
