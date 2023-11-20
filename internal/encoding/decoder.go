package encoding

type (
	// Decoder is the interface for decoding data.
	Decoder[S, T any] interface {
		Decode(source S, target T) error
	}

	DecoderFunc[S, T any] func(source S, target T) error
)

var _ Decoder[any, any] = DecoderFunc[any, any](func(source, target any) error { return nil })

func (dec DecoderFunc[S, T]) Decode(source S, target T) error {
	return dec(source, target)
}
