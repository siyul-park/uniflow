package primitive

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/internal/encoding"
	"github.com/siyul-park/uniflow/internal/util"
)

// NewPointerEncoder is encode *T to T.
func NewPointerEncoder(encoder encoding.Encoder[any, Object]) encoding.Encoder[any, Object] {
	return encoding.EncoderFunc[any, Object](func(source any) (Object, error) {
		if util.IsNil(source) {
			return nil, nil
		}
		if s := reflect.ValueOf(source); s.Kind() == reflect.Pointer {
			return encoder.Encode(s.Elem().Interface())
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

// NewPointerDecoder is decode T to *T.
func NewPointerDecoder(decoder encoding.Decoder[Object, any]) encoding.Decoder[Object, any] {
	return encoding.DecoderFunc[Object, any](func(source Object, target any) error {
		if util.IsNil(source) {
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
