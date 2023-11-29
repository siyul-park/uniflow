package primitive

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

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
