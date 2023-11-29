package primitive

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
)

// NewPointerEncoder is encode Object to Object.
func NewShortcutEncoder() encoding.Encoder[any, Value] {
	return encoding.EncoderFunc[any, Value](func(source any) (Value, error) {
		if s, ok := source.(Value); ok {
			return s, nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

// NewShortcutDecoder is decode Object to Object.
func NewShortcutDecoder() encoding.Decoder[Value, any] {
	return encoding.DecoderFunc[Value, any](func(source Value, target any) error {
		if t, ok := target.(*Value); ok {
			*t = source
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
