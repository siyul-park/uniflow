package primitive

import (
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/internal/encoding"
)

// NewPointerEncoder is encode Object to Object.
func NewShortcutEncoder() encoding.Encoder[any, Object] {
	return encoding.EncoderFunc[any, Object](func(source any) (Object, error) {
		if s, ok := source.(Object); ok {
			return s, nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

// NewShortcutDecoder is decode Object to Object.
func NewShortcutDecoder() encoding.Decoder[Object, any] {
	return encoding.DecoderFunc[Object, any](func(source Object, target any) error {
		if t, ok := target.(*Object); ok {
			*t = source
			return nil
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
