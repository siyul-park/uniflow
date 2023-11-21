package primitive

import (
	"reflect"

	"github.com/siyul-park/uniflow/internal/encoding"
)

var (
	textEncoder   = encoding.NewEncoderGroup[any, Object]()
	binaryEncoder = encoding.NewEncoderGroup[any, Object]()
	decoder       = encoding.NewDecoderGroup[Object, any]()
)

var (
	typeAny = reflect.TypeOf((*any)(nil)).Elem()
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
func MarshalText(v any) (Object, error) {
	return textEncoder.Encode(v)
}

// MarshalBinary returns the Object of v.
func MarshalBinary(v any) (Object, error) {
	return binaryEncoder.Encode(v)
}

// Unmarshal parses the Object and stores the result.
func Unmarshal(data Object, v any) error {
	return decoder.Decode(data, v)
}