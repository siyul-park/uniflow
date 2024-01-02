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
	textEncoder.Add(NewShortcutEncoder())
	textEncoder.Add(NewBoolEncoder())
	textEncoder.Add(NewFloatEncoder())
	textEncoder.Add(NewIntEncoder())
	textEncoder.Add(NewUintEncoder())
	textEncoder.Add(NewStringEncoder())
	textEncoder.Add(NewBinaryEncoder())
	textEncoder.Add(NewSliceEncoder(textEncoder))
	textEncoder.Add(NewMapEncoder(textEncoder))
	textEncoder.Add(NewPointerEncoder(textEncoder))

	binaryEncoder.Add(NewShortcutEncoder())
	binaryEncoder.Add(NewBoolEncoder())
	binaryEncoder.Add(NewFloatEncoder())
	binaryEncoder.Add(NewIntEncoder())
	binaryEncoder.Add(NewUintEncoder())
	binaryEncoder.Add(NewBinaryEncoder())
	binaryEncoder.Add(NewStringEncoder())
	binaryEncoder.Add(NewSliceEncoder(binaryEncoder))
	binaryEncoder.Add(NewMapEncoder(binaryEncoder))
	binaryEncoder.Add(NewPointerEncoder(binaryEncoder))

	decoder.Add(NewShortcutDecoder())
	decoder.Add(NewBoolDecoder())
	decoder.Add(NewFloatDecoder())
	decoder.Add(NewIntDecoder())
	decoder.Add(NewUintDecoder())
	decoder.Add(NewStringDecoder())
	decoder.Add(NewBinaryDecoder())
	decoder.Add(NewSliceDecoder(decoder))
	decoder.Add(NewMapDecoder(decoder))
	decoder.Add(NewPointerDecoder(decoder))
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

// NewPointerEncoder is encode *T to T.
func NewPointerEncoder(encoder encoding.Encoder[any, Value]) encoding.Encoder[any, Value] {
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

// NewPointerDecoder is decode T to *T.
func NewPointerDecoder(decoder encoding.Decoder[Value, any]) encoding.Decoder[Value, any] {
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

// NewPointerEncoder is encode Value to Value.
func NewShortcutEncoder() encoding.Encoder[any, Value] {
	return encoding.EncoderFunc[any, Value](func(source any) (Value, error) {
		if s, ok := source.(Value); ok {
			return s, nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

// NewShortcutDecoder is decode Value to Value.
func NewShortcutDecoder() encoding.Decoder[Value, any] {
	return encoding.DecoderFunc[Value, any](func(source Value, target any) error {
		if t, ok := target.(*Value); ok {
			*t = source
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
