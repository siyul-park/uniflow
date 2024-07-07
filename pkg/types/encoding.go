package types

import (
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

type Marshaler interface {
	MarshalObject() (Value, error)
}

type Unmarshaler interface {
	UnmarshalObject(Value) error
}

var (
	TextEncoder   = encoding.NewEncodeAssembler[any, Value]()
	BinaryEncoder = encoding.NewEncodeAssembler[any, Value]()
	Decoder       = encoding.NewDecodeAssembler[Value, any]()
)

func init() {
	TextEncoder.Add(newShortcutEncoder())
	TextEncoder.Add(newExpandedEncoder())
	TextEncoder.Add(newErrorEncoder())
	TextEncoder.Add(newStringEncoder())
	TextEncoder.Add(newBinaryEncoder())
	TextEncoder.Add(newBooleanEncoder())
	TextEncoder.Add(newFloatEncoder())
	TextEncoder.Add(NewIntegerEncoder())
	TextEncoder.Add(newUintegerEncoder())
	TextEncoder.Add(newSliceEncoder(TextEncoder))
	TextEncoder.Add(newMapEncoder(TextEncoder))
	TextEncoder.Add(newPointerEncoder(TextEncoder))

	BinaryEncoder.Add(newShortcutEncoder())
	BinaryEncoder.Add(newExpandedEncoder())
	BinaryEncoder.Add(newErrorEncoder())
	BinaryEncoder.Add(newBinaryEncoder())
	BinaryEncoder.Add(newStringEncoder())
	BinaryEncoder.Add(newBooleanEncoder())
	BinaryEncoder.Add(newFloatEncoder())
	BinaryEncoder.Add(NewIntegerEncoder())
	BinaryEncoder.Add(newUintegerEncoder())
	BinaryEncoder.Add(newSliceEncoder(BinaryEncoder))
	BinaryEncoder.Add(newMapEncoder(BinaryEncoder))
	BinaryEncoder.Add(newPointerEncoder(BinaryEncoder))

	Decoder.Add(newShortcutDecoder())
	Decoder.Add(newExpandedDecoder())
	Decoder.Add(newErrorDecoder())
	Decoder.Add(newBinaryDecoder())
	Decoder.Add(newStringDecoder())
	Decoder.Add(newBooleanDecoder())
	Decoder.Add(newFloatDecoder())
	Decoder.Add(NewIntegerDecoder())
	Decoder.Add(newUintegerDecoder())
	Decoder.Add(newSliceDecoder(Decoder))
	Decoder.Add(newMapDecoder(Decoder))
	Decoder.Add(newPointerDecoder(Decoder))
}

func newShortcutEncoder() encoding.EncodeCompiler[any, Value] {
	typeValue := reflect.TypeOf((*Value)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ != nil && typ.ConvertibleTo(typeValue) {
			return encoding.EncodeFunc[any, Value](func(source any) (Value, error) {
				s := source.(Value)
				return s, nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}

func newShortcutDecoder() encoding.DecodeCompiler[Value] {
	typeValue := reflect.TypeOf((*Value)(nil)).Elem()

	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer && typ.Elem().ConvertibleTo(typeValue) {
			return encoding.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				*(*Value)(target) = source
				return nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}

func newExpandedEncoder() encoding.EncodeCompiler[any, Value] {
	typeMarshaler := reflect.TypeOf((*Marshaler)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ != nil && typ.Kind() == reflect.Pointer && typ.ConvertibleTo(typeMarshaler) {
			return encoding.EncodeFunc[any, Value](func(source any) (Value, error) {
				s := source.(Marshaler)
				return s.MarshalObject()
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}

func newExpandedDecoder() encoding.DecodeCompiler[Value] {
	typeUnmarshaler := reflect.TypeOf((*Unmarshaler)(nil)).Elem()

	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.ConvertibleTo(typeUnmarshaler) {
			return encoding.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target).Interface().(Unmarshaler)
				return t.UnmarshalObject(source)
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}

func newPointerEncoder(encoder *encoding.EncodeAssembler[any, Value]) encoding.EncodeCompiler[any, Value] {
	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ == nil {
			return encoding.EncodeFunc[any, Value](func(source any) (Value, error) {
				return nil, nil
			}), nil
		} else if typ.Kind() == reflect.Pointer {
			enc, err := encoder.Compile(typ.Elem())
			if err != nil {
				return nil, err
			}

			return encoding.EncodeFunc[any, Value](func(source any) (Value, error) {
				if source == nil {
					return nil, nil
				}
				s := reflect.ValueOf(source)
				return enc.Encode(s.Elem().Interface())
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}

func newPointerDecoder(decoder *encoding.DecodeAssembler[Value, any]) encoding.DecodeCompiler[Value] {
	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.Pointer {
			dec, err := decoder.Compile(typ.Elem())
			if err != nil {
				return nil, err
			}

			return encoding.DecodeFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target)
				if t.Elem().IsNil() {
					zero := reflect.New(t.Type().Elem().Elem())
					t.Elem().Set(zero)
				}
				return dec.Decode(source, t.Elem().UnsafePointer())
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrInvalidArgument)
	})
}
