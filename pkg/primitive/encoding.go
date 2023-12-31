package primitive

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

var (
	textEncoder   = encoding.NewEncoderGroup[any, Value]()
	binaryEncoder = encoding.NewEncoderGroup[any, Value]()
	decoder       = encoding.NewDecoderGroup[Value, any]()
)

func init() {
	textEncoder.Add(newShortcutEncoder())
	textEncoder.Add(newBoolEncoder())
	textEncoder.Add(newFloatEncoder())
	textEncoder.Add(newIntEncoder())
	textEncoder.Add(newUintEncoder())
	textEncoder.Add(newStringEncoder())
	textEncoder.Add(newBinaryEncoder())
	textEncoder.Add(newSliceEncoder(textEncoder))
	textEncoder.Add(newMapEncoder(textEncoder))
	textEncoder.Add(newPointerEncoder(textEncoder))

	binaryEncoder.Add(newShortcutEncoder())
	binaryEncoder.Add(newBoolEncoder())
	binaryEncoder.Add(newFloatEncoder())
	binaryEncoder.Add(newIntEncoder())
	binaryEncoder.Add(newUintEncoder())
	binaryEncoder.Add(newBinaryEncoder())
	binaryEncoder.Add(newStringEncoder())
	binaryEncoder.Add(newSliceEncoder(binaryEncoder))
	binaryEncoder.Add(newMapEncoder(binaryEncoder))
	binaryEncoder.Add(newPointerEncoder(binaryEncoder))

	decoder.Add(newShortcutDecoder())
	decoder.Add(newBoolDecoder())
	decoder.Add(newFloatDecoder())
	decoder.Add(newIntDecoder())
	decoder.Add(newUintDecoder())
	decoder.Add(newStringDecoder())
	decoder.Add(newBinaryDecoder())
	decoder.Add(newSliceDecoder(decoder))
	decoder.Add(newMapDecoder(decoder))
	decoder.Add(newPointerDecoder(decoder))
}

// MarshalText returns the Object of v.
func MarshalText(v any) (Value, error) {
	return textEncoder.Encode(v)
}

// MarshalBinary returns the Object of v.
func MarshalBinary(v any) (Value, error) {
	return binaryEncoder.Encode(v)
}

// Unmarshal parses the Object and stores the result.
func Unmarshal(data Value, v any) error {
	return decoder.Decode(data, v)
}

func newPointerEncoder(encoder encoding.Encoder[any, Value]) encoding.Encoder[any, Value] {
	return encoding.EncoderFunc[any, Value](func(source any) (Value, error) {
		if source == nil {
			return nil, nil
		}
		if s := reflect.ValueOf(source); s.Kind() == reflect.Pointer {
			return encoder.Encode(s.Elem().Interface())
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newPointerDecoder(decoder encoding.Decoder[Value, any]) encoding.Decoder[Value, any] {
	return encoding.DecoderFunc[Value, any](func(source Value, target any) error {
		if source == nil {
			return nil
		}
		if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Pointer {
			if t.Elem().IsNil() {
				zero := reflect.New(t.Type().Elem().Elem())
				t.Elem().Set(zero)
			}
			return decoder.Decode(source, t.Elem().Interface())
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newShortcutEncoder() encoding.Encoder[any, Value] {
	return encoding.EncoderFunc[any, Value](func(source any) (Value, error) {
		if s, ok := source.(Value); ok {
			return s, nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

func newShortcutDecoder() encoding.Decoder[Value, any] {
	return encoding.DecoderFunc[Value, any](func(source Value, target any) error {
		if t, ok := target.(*Value); ok {
			*t = source
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
