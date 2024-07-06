package types

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
	textEncoder.Add(newBooleanEncoder())
	textEncoder.Add(newFloatEncoder())
	textEncoder.Add(NewIntegerEncoder())
	textEncoder.Add(newUintegerEncoder())
	textEncoder.Add(newSliceEncoder(textEncoder))
	textEncoder.Add(newMapEncoder(textEncoder))
	textEncoder.Add(newPointerEncoder(textEncoder))

	binaryEncoder.Add(newShortcutEncoder())
	binaryEncoder.Add(newExpandedEncoder())
	binaryEncoder.Add(newErrorEncoder())
	binaryEncoder.Add(newBinaryEncoder())
	binaryEncoder.Add(newStringEncoder())
	binaryEncoder.Add(newBooleanEncoder())
	binaryEncoder.Add(newFloatEncoder())
	binaryEncoder.Add(NewIntegerEncoder())
	binaryEncoder.Add(newUintegerEncoder())
	binaryEncoder.Add(newSliceEncoder(binaryEncoder))
	binaryEncoder.Add(newMapEncoder(binaryEncoder))
	binaryEncoder.Add(newPointerEncoder(binaryEncoder))

	decoder.Add(newShortcutDecoder())
	decoder.Add(newExpandedDecoder())
	decoder.Add(newErrorDecoder())
	decoder.Add(newBinaryDecoder())
	decoder.Add(newStringDecoder())
	decoder.Add(newBooleanDecoder())
	decoder.Add(newFloatDecoder())
	decoder.Add(NewIntegerDecoder())
	decoder.Add(newUintegerDecoder())
	decoder.Add(newSliceDecoder(decoder))
	decoder.Add(newMapDecoder(decoder))
	decoder.Add(newPointerDecoder(decoder))
}

// MarshalText returns the Object of v.
func MarshalText(v any) (Object, error) {
	if v == nil {
		return nil, nil
	}
	return textEncoder.Encode(v)
}

// MarshalBinary returns the Object of v.
func MarshalBinary(v any) (Object, error) {
	if v == nil {
		return nil, nil
	}
	return binaryEncoder.Encode(v)
}

// Unmarshal parses the Object and stores the result.
func Unmarshal(data Object, v any) error {
	return decoder.Decode(data, v)
}

func newShortcutEncoder() encoding.EncodeCompiler[any, Object] {
	typeValue := reflect.TypeOf((*Object)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, Object](func(typ reflect.Type) (encoding.Encoder[any, Object], error) {
		if typ.ConvertibleTo(typeValue) {
			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				s := source.(Object)
				return s, nil
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

func newExpandedEncoder() encoding.EncodeCompiler[any, Object] {
	typeMarshaler := reflect.TypeOf((*Marshaler)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, Object](func(typ reflect.Type) (encoding.Encoder[any, Object], error) {
		if typ.Kind() == reflect.Pointer && typ.ConvertibleTo(typeMarshaler) {
			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				s := source.(Marshaler)
				return s.MarshalObject()
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

func newPointerEncoder(encoder *encoding.EncodeAssembler[any, Object]) encoding.EncodeCompiler[any, Object] {
	return encoding.EncodeCompilerFunc[any, Object](func(typ reflect.Type) (encoding.Encoder[any, Object], error) {
		if typ.Kind() == reflect.Pointer {
			enc, err := encoder.Compile(typ.Elem())
			if err != nil {
				return nil, err
			}

			return encoding.EncodeFunc[any, Object](func(source any) (Object, error) {
				s := reflect.ValueOf(source)
				return enc.Encode(s.Elem().Interface())
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
