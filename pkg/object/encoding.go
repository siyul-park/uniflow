package object

import (
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

type Marshaler interface {
	MarshalObject() (Object, error)
}

type Unmarshaler interface {
	UnmarshalObject(Object) error
}

var (
	textEncoder   = encoding.NewAssembler[*Object, any]()
	binaryEncoder = encoding.NewAssembler[*Object, any]()
	decoder       = encoding.NewAssembler[Object, any]()
)

func init() {
	textEncoder.Add(newShortcutEncoder())
	textEncoder.Add(newExpandedEncoder())
	textEncoder.Add(newStringEncoder())
	textEncoder.Add(newBinaryEncoder())
	textEncoder.Add(newBoolEncoder())
	textEncoder.Add(newFloatEncoder())
	textEncoder.Add(newIntegerEncoder())
	textEncoder.Add(newUintegerEncoder())
	textEncoder.Add(newSliceEncoder(textEncoder))
	textEncoder.Add(newMapEncoder(textEncoder))
	textEncoder.Add(newPointerEncoder(textEncoder))

	binaryEncoder.Add(newShortcutEncoder())
	binaryEncoder.Add(newExpandedEncoder())
	binaryEncoder.Add(newBinaryEncoder())
	binaryEncoder.Add(newStringEncoder())
	binaryEncoder.Add(newBoolEncoder())
	binaryEncoder.Add(newFloatEncoder())
	binaryEncoder.Add(newIntegerEncoder())
	binaryEncoder.Add(newUintegerEncoder())
	binaryEncoder.Add(newSliceEncoder(binaryEncoder))
	binaryEncoder.Add(newMapEncoder(binaryEncoder))
	binaryEncoder.Add(newPointerEncoder(binaryEncoder))

	decoder.Add(newShortcutDecoder())
	decoder.Add(newExpandedDecoder())
	decoder.Add(newBinaryDecoder())
	decoder.Add(newStringDecoder())
	decoder.Add(newBoolDecoder())
	decoder.Add(newFloatDecoder())
	decoder.Add(newIntegerDecoder())
	decoder.Add(newUintegerDecoder())
	decoder.Add(newSliceDecoder(decoder))
	decoder.Add(newMapDecoder(decoder))
	decoder.Add(newPointerDecoder(decoder))
}

// MarshalText returns the Object of v.
func MarshalText(v any) (Object, error) {
	var data Object
	if err := textEncoder.Encode(&data, v); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

// MarshalBinary returns the Object of v.
func MarshalBinary(v any) (Object, error) {
	var data Object
	if err := binaryEncoder.Encode(&data, v); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

// Unmarshal parses the Object and stores the result.
func Unmarshal(data Object, v any) error {
	return decoder.Encode(data, v)
}

func newShortcutEncoder() encoding.Compiler[*Object] {
	typeValue := reflect.TypeOf((*Object)(nil)).Elem()

	return encoding.CompilerFunc[*Object](func(typ reflect.Type) (encoding.Encoder[*Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().ConvertibleTo(typeValue) {
			return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target).Elem().Interface().(Object)
				*source = t
				return nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newShortcutDecoder() encoding.Compiler[Object] {
	typeValue := reflect.TypeOf((*Object)(nil)).Elem()

	return encoding.CompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().ConvertibleTo(typeValue) {
			return encoding.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
				*(*Object)(target) = source
				return nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newExpandedEncoder() encoding.Compiler[*Object] {
	typeMarshaler := reflect.TypeOf((*Marshaler)(nil)).Elem()

	return encoding.CompilerFunc[*Object](func(typ reflect.Type) (encoding.Encoder[*Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.ConvertibleTo(typeMarshaler) {
			return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target).Interface().(Marshaler)
				if s, err := t.MarshalObject(); err != nil {
					return err
				} else {
					*source = s
				}
				return nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newExpandedDecoder() encoding.Compiler[Object] {
	typeUnmarshaler := reflect.TypeOf((*Unmarshaler)(nil)).Elem()

	return encoding.CompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[Object, unsafe.Pointer], error) {
		if typ.ConvertibleTo(typeUnmarshaler) {
			return encoding.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target).Interface().(Unmarshaler)
				return t.UnmarshalObject(source)
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newPointerEncoder(encoder *encoding.Assembler[*Object, any]) encoding.Compiler[*Object] {
	return encoding.CompilerFunc[*Object](func(typ reflect.Type) (encoding.Encoder[*Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.Pointer {
			return encoding.EncodeFunc[*Object, unsafe.Pointer](func(source *Object, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target)
				return encoder.Encode(source, t.Elem().Interface())
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newPointerDecoder(decoder *encoding.Assembler[Object, any]) encoding.Compiler[Object] {
	return encoding.CompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.Pointer {
			return encoding.EncodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target)
				if t.Elem().IsNil() {
					zero := reflect.New(t.Type().Elem().Elem())
					t.Elem().Set(zero)
				}
				return decoder.Encode(source, t.Elem().Interface())
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
