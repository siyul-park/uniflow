package types

import (
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

var (
	// Encoder is a global encoding assembler used to encode values into the custom Value type.
	Encoder = encoding.NewEncodeAssembler[any, Value]()
	// Decoder is a global decoding assembler used to decode values from the custom Value type.
	Decoder = encoding.NewDecodeAssembler[Value, any]()
)

func init() {
	Encoder.Add(newPointerEncoder(Encoder))
	Encoder.Add(newMapEncoder(Encoder))
	Encoder.Add(newSliceEncoder(Encoder))
	Encoder.Add(newJSONEncoder(Encoder))
	Encoder.Add(newUintegerEncoder())
	Encoder.Add(newIntegerEncoder())
	Encoder.Add(newFloatEncoder())
	Encoder.Add(newBooleanEncoder())
	Encoder.Add(newBufferEncoder())
	Encoder.Add(newBinaryEncoder())
	Encoder.Add(newStringEncoder())
	Encoder.Add(newErrorEncoder())
	Encoder.Add(newTimeEncoder())
	Encoder.Add(newDurationEncoder())
	Encoder.Add(newShortcutEncoder())

	Decoder.Add(newPointerDecoder(Decoder))
	Decoder.Add(newMapDecoder(Decoder))
	Decoder.Add(newSliceDecoder(Decoder))
	Decoder.Add(newJSONDecoder(Decoder))
	Decoder.Add(newUintegerDecoder())
	Decoder.Add(newIntegerDecoder())
	Decoder.Add(newFloatDecoder())
	Decoder.Add(newBooleanDecoder())
	Decoder.Add(newBufferDecoder())
	Decoder.Add(newBinaryDecoder())
	Decoder.Add(newStringDecoder())
	Decoder.Add(newErrorDecoder())
	Decoder.Add(newTimeDecoder())
	Decoder.Add(newDurationDecoder())
	Decoder.Add(newShortcutDecoder())
}

// Marshal encodes the given value into a Value using the global Encoder.
func Marshal(val any) (Value, error) {
	return Encoder.Encode(val)
}

// Unmarshal decodes the given Value into the provided target using the global Decoder.
func Unmarshal(data Value, v any) error {
	return Decoder.Decode(data, v)
}

func newShortcutEncoder() encoding.EncodeCompiler[any, Value] {
	typeValue := reflect.TypeOf((*Value)(nil)).Elem()

	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ != nil && typ.ConvertibleTo(typeValue) {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				s := source.(Value)
				return s, nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newShortcutDecoder() encoding.DecodeCompiler[Value] {
	typeValue := reflect.TypeOf((*Value)(nil)).Elem()

	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ != nil && typ.Kind() == reflect.Pointer && typ.Elem().ConvertibleTo(typeValue) {
			return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
				s := reflect.ValueOf(source)
				t := reflect.NewAt(typ.Elem(), target).Elem()
				if s.Type().ConvertibleTo(typ.Elem()) {
					t.Set(s)
					return nil
				}
				return errors.WithStack(encoding.ErrUnsupportedType)
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newPointerEncoder(encoder *encoding.EncodeAssembler[any, Value]) encoding.EncodeCompiler[any, Value] {
	return encoding.EncodeCompilerFunc[any, Value](func(typ reflect.Type) (encoding.Encoder[any, Value], error) {
		if typ == nil {
			return encoding.EncodeFunc(func(source any) (Value, error) {
				return nil, nil
			}), nil
		} else if typ.Kind() == reflect.Pointer {
			enc, err := encoder.Compile(typ.Elem())
			if err != nil {
				return nil, err
			}

			return encoding.EncodeFunc(func(source any) (Value, error) {
				if source == nil {
					return nil, nil
				}
				s := reflect.ValueOf(source)
				return enc.Encode(s.Elem().Interface())
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}

func newPointerDecoder(decoder *encoding.DecodeAssembler[Value, any]) encoding.DecodeCompiler[Value] {
	return encoding.DecodeCompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ == nil {
			return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
				return nil
			}), nil
		} else if typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.Pointer {
			dec, err := decoder.Compile(typ.Elem())
			if err != nil {
				return nil, err
			}

			return encoding.DecodeFunc(func(source Value, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target)
				if t.Elem().IsNil() {
					zero := reflect.New(t.Type().Elem().Elem())
					t.Elem().Set(zero)
				}
				return dec.Decode(source, t.Elem().UnsafePointer())
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	})
}
