package encoding

// Encoder is an interface for encoding data.
type Encoder[S, T any] interface {
	// Encode encodes data from the source to the target type.
	Encode(source S) (T, error)
}

// EncodeFunc is a function type that implements the Encoder interface.
type EncodeFunc[S, T any] func(source S) (T, error)

var _ Encoder[any, any] = (EncodeFunc[any, any])(nil)

// Encode calls the underlying function to perform encoding.
func (f EncodeFunc[S, T]) Encode(source S) (T, error) {
	return f(source)
}
