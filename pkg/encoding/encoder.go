package encoding

type (
	// Encoder is an interface for encoding data.
	Encoder[S, T any] interface {
		Encode(source S) (T, error)
	}

	EncoderFunc[S, T any] func(source S) (T, error)
)

var _ Encoder[any, any] = EncoderFunc[any, any](func(_ any) (any, error) { return nil, nil })

func (enc EncoderFunc[S, T]) Encode(source S) (T, error) {
	return enc(source)
}
