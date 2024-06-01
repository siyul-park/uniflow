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
	textEncoder   = encoding.NewEncodeAssembler[any, Object]()
	binaryEncoder = encoding.NewEncodeAssembler[any, Object]()
	decoder       = encoding.NewDecodeAssembler[Object, any]()
)

func init() {
	textEncoder.Add(newShortcutEncoder())
	textEncoder.Add(newExpandedEncoder())
	textEncoder.Add(newErrorEncoder())
	textEncoder.Add(newStringEncoder())
	textEncoder.Add(newBinaryEncoder())
	textEncoder.Add(newBoolEncoder())
	textEncoder.Add(newFloatEncoder())
	textEncoder.Add(NewIntEncoder())
	textEncoder.Add(newUintEncoder())
	textEncoder.Add(newSliceEncoder(textEncoder))
	textEncoder.Add(newMapEncoder(textEncoder))
	textEncoder.Add(newPointerEncoder(textEncoder))

	binaryEncoder.Add(newShortcutEncoder())
	binaryEncoder.Add(newExpandedEncoder())
	binaryEncoder.Add(newErrorEncoder())
	binaryEncoder.Add(newBinaryEncoder())
	binaryEncoder.Add(newStringEncoder())
	binaryEncoder.Add(newBoolEncoder())
	binaryEncoder.Add(newFloatEncoder())
	binaryEncoder.Add(NewIntEncoder())
	binaryEncoder.Add(newUintEncoder())
	binaryEncoder.Add(newSliceEncoder(binaryEncoder))
	binaryEncoder.Add(newMapEncoder(binaryEncoder))
	binaryEncoder.Add(newPointerEncoder(binaryEncoder))

	decoder.Add(newShortcutDecoder())
	decoder.Add(newExpandedDecoder())
	decoder.Add(newErrorDecoder())
	decoder.Add(newBinaryDecoder())
	decoder.Add(newStringDecoder())
	decoder.Add(newBoolDecoder())
	decoder.Add(newFloatDecoder())
	decoder.Add(NewIntDecoder())
	decoder.Add(newUintDecoder())
	decoder.Add(newSliceDecoder(decoder))
	decoder.Add(newMapDecoder(decoder))
	decoder.Add(newPointerDecoder(decoder))
}

// MarshalText returns the Object of v.
func MarshalText(v any) (Object, error) {
	if v == nil {
		return nil, nil
	}
	ptr := reflect.New(reflect.TypeOf(v))
	ptr.Elem().Set(reflect.ValueOf(v))
	return textEncoder.Encode(ptr.Interface())
}

// MarshalBinary returns the Object of v.
func MarshalBinary(v any) (Object, error) {
	if v == nil {
		return nil, nil
	}
	ptr := reflect.New(reflect.TypeOf(v))
	ptr.Elem().Set(reflect.ValueOf(v))
	return binaryEncoder.Encode(ptr.Interface())
}

// Unmarshal parses the Object and stores the result.
func Unmarshal(data Object, v any) error {
	return decoder.Decode(data, v)
}

func newShortcutEncoder() encoding.EncodeCompiler[Object] {
	typeValue := reflect.TypeOf((*Object)(nil)).Elem()

	return encoding.EncodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[unsafe.Pointer, Object], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().ConvertibleTo(typeValue) {
			return encoding.EncodeFunc[unsafe.Pointer, Object](func(source unsafe.Pointer) (Object, error) {
				t := reflect.NewAt(typ.Elem(), source).Elem().Interface().(Object)
				return t, nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newShortcutDecoder() encoding.DecodeCompiler[Object] {
	typeValue := reflect.TypeOf((*Object)(nil)).Elem()

	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().ConvertibleTo(typeValue) {
			return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
				*(*Object)(target) = source
				return nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newExpandedEncoder() encoding.EncodeCompiler[Object] {
	typeMarshaler := reflect.TypeOf((*Marshaler)(nil)).Elem()

	return encoding.EncodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[unsafe.Pointer, Object], error) {
		if typ.Kind() == reflect.Pointer && typ.ConvertibleTo(typeMarshaler) {
			return encoding.EncodeFunc[unsafe.Pointer, Object](func(target unsafe.Pointer) (Object, error) {
				t := reflect.NewAt(typ.Elem(), target).Interface().(Marshaler)
				if s, err := t.MarshalObject(); err != nil {
					return nil, err
				} else {
					return s, nil
				}
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newExpandedDecoder() encoding.DecodeCompiler[Object] {
	typeUnmarshaler := reflect.TypeOf((*Unmarshaler)(nil)).Elem()

	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.ConvertibleTo(typeUnmarshaler) {
			return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target).Interface().(Unmarshaler)
				return t.UnmarshalObject(source)
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newPointerEncoder(encoder *encoding.EncodeAssembler[any, Object]) encoding.EncodeCompiler[Object] {
	return encoding.EncodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Encoder[unsafe.Pointer, Object], error) {
		if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Ptr {
			enc, err := encoder.Compile(typ.Elem())
			if err != nil {
				return nil, err
			}

			return encoding.EncodeFunc[unsafe.Pointer, Object](func(target unsafe.Pointer) (Object, error) {
				t := reflect.NewAt(typ.Elem(), target)
				return enc.Encode(t.Elem().UnsafePointer())
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newPointerDecoder(decoder *encoding.DecodeAssembler[Object, any]) encoding.DecodeCompiler[Object] {
	return encoding.DecodeCompilerFunc[Object](func(typ reflect.Type) (encoding.Decoder[Object, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.Pointer {
			dec, err := decoder.Compile(typ.Elem())
			if err != nil {
				return nil, err
			}

			return encoding.DecodeFunc[Object, unsafe.Pointer](func(source Object, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target)
				if t.Elem().IsNil() {
					zero := reflect.New(t.Type().Elem().Elem())
					t.Elem().Set(zero)
				}
				return dec.Decode(source, t.Elem().UnsafePointer())
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
