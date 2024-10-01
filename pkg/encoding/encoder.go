package encoding

// Encoder is an interface for encoding data.
type Encoder[S, T any] interface {
	// Encode encodes data from the source to the target type.
	Encode(source S) (T, error)
}

type encoder[S, T any] struct {
	encode func(source S) (T, error)
}

var _ Encoder[any, any] = (*encoder[any, any])(nil)

// EncodeFunc creates an Encoder instance from a provided function.
func EncodeFunc[S, T any](encode func(source S) (T, error)) Encoder[S, T] {
	return &encoder[S, T]{encode: encode}
}

// Encode calls the underlying encode function.
func (e *encoder[S, T]) Encode(source S) (T, error) {
	return e.encode(source)
}
