package primitive

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"reflect"
	"unsafe"
)

type Marshaler interface {
	MarshalPrimitive() (Value, error)
}

type Unmarshaler interface {
	UnmarshalPrimitive(Value) error
}

var (
	textEncoder   = encoding.NewCompiledDecoder[*Value, any]()
	binaryEncoder = encoding.NewCompiledDecoder[*Value, any]()
	decoder       = encoding.NewCompiledDecoder[Value, any]()
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
func MarshalText(v any) (Value, error) {
	var data Value
	if err := textEncoder.Decode(&data, v); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

// MarshalBinary returns the Object of v.
func MarshalBinary(v any) (Value, error) {
	var data Value
	if err := binaryEncoder.Decode(&data, v); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

// Unmarshal parses the Object and stores the result.
func Unmarshal(data Value, v any) error {
	return decoder.Decode(data, v)
}

func newShortcutEncoder() encoding.Compiler[*Value] {
	typeValue := reflect.TypeOf((*Value)(nil)).Elem()

	return encoding.CompilerFunc[*Value](func(typ reflect.Type) (encoding.Decoder[*Value, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().ConvertibleTo(typeValue) {
			return encoding.DecoderFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target).Elem().Interface().(Value)
				*source = t
				return nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newShortcutDecoder() encoding.Compiler[Value] {
	typeValue := reflect.TypeOf((*Value)(nil)).Elem()

	return encoding.CompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().ConvertibleTo(typeValue) {
			return encoding.DecoderFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				*(*Value)(target) = source
				return nil
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newExpandedEncoder() encoding.Compiler[*Value] {
	typeMarshaler := reflect.TypeOf((*Marshaler)(nil)).Elem()

	return encoding.CompilerFunc[*Value](func(typ reflect.Type) (encoding.Decoder[*Value, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.ConvertibleTo(typeMarshaler) {
			return encoding.DecoderFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target).Interface().(Marshaler)
				if s, err := t.MarshalPrimitive(); err != nil {
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

func newExpandedDecoder() encoding.Compiler[Value] {
	typeUnmarshaler := reflect.TypeOf((*Unmarshaler)(nil)).Elem()

	return encoding.CompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ.ConvertibleTo(typeUnmarshaler) {
			return encoding.DecoderFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target).Interface().(Unmarshaler)
				return t.UnmarshalPrimitive(source)
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newPointerEncoder(encoder *encoding.CompiledDecoder[*Value, any]) encoding.Compiler[*Value] {
	return encoding.CompilerFunc[*Value](func(typ reflect.Type) (encoding.Decoder[*Value, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.Pointer {
			return encoding.DecoderFunc[*Value, unsafe.Pointer](func(source *Value, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target)
				return encoder.Decode(source, t.Elem().Interface())
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newPointerDecoder(decoder *encoding.CompiledDecoder[Value, any]) encoding.Compiler[Value] {
	return encoding.CompilerFunc[Value](func(typ reflect.Type) (encoding.Decoder[Value, unsafe.Pointer], error) {
		if typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.Pointer {
			return encoding.DecoderFunc[Value, unsafe.Pointer](func(source Value, target unsafe.Pointer) error {
				t := reflect.NewAt(typ.Elem(), target)
				if t.Elem().IsNil() {
					zero := reflect.New(t.Type().Elem().Elem())
					t.Elem().Set(zero)
				}
				return decoder.Decode(source, t.Elem().Interface())
			}), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
