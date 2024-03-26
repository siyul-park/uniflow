package primitive

import (
	"github.com/siyul-park/uniflow/pkg/encoding"
	"reflect"
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

var typeAny = reflect.TypeOf((*any)(nil)).Elem()

func init() {
	textEncoder.Add(newBinaryEncoder())
	binaryEncoder.Add(newBinaryEncoder())
	decoder.Add(newBinaryDecoder())
}

// MarshalText returns the Object of v.
func MarshalText(v any) (Value, error) {
	var data Value
	if err := textEncoder.Decode(&data, &v); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

// MarshalBinary returns the Object of v.
func MarshalBinary(v any) (Value, error) {
	var data Value
	if err := binaryEncoder.Decode(&data, &v); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

// Unmarshal parses the Object and stores the result.
func Unmarshal(data Value, v any) error {
	return decoder.Decode(data, v)
}
