package encoding

// Decoder is the interface for decoding data.
type Decoder[S, T any] interface {
	Decode(source S, target T) error
}

// DecoderFunc is a function type that implements the Decoder interface.
type DecoderFunc[S, T any] func(source S, target T) error

var _ Decoder[any, any] = DecoderFunc[any, any](func(source, target any) error { return nil })

// Decode calls the underlying function to perform decoding.
func (dec DecoderFunc[S, T]) Decode(source S, target T) error {
	return dec(source, target)
}
