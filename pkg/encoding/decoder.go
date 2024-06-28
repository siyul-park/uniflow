package encoding

// Decoder is an interface for decoding data.
type Decoder[S, T any] interface {
	// Decode decodes data from the source to the target.
	Decode(source S, target T) error
}

// DecodeFunc is a function type that implements the Decoder interface.
type DecodeFunc[S, T any] func(source S, target T) error

var _ Decoder[any, any] = (DecodeFunc[any, any])(nil)

// Decode calls the underlying function to perform decoding.
func (f DecodeFunc[S, T]) Decode(source S, target T) error {
	return f(source, target)
}
