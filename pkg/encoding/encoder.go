package encoding

// Encoder is an interface for encoding data.
type Encoder[S, T any] interface {
	// Encode encodes data from the source to the target.
	Encode(source S) (T, error)
}

// EncoderFunc is a function type that implements the Encoder interface.
type EncoderFunc[S, T any] func(source S) (T, error)

var _ Encoder[any, any] = EncoderFunc[any, any](func(_ any) (any, error) { return nil, nil })

// Encode calls the underlying function to perform encoding.
func (enc EncoderFunc[S, T]) Encode(source S) (T, error) {
	return enc(source)
}
