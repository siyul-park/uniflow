package encoding

// Encoder is an interface for encoding data.
type Encoder[S, T any] interface {
	// Encode encodes data from the source to the target.
	Encode(source S, target T) error
}

// EncodeFunc is a function type that implements the Encoder interface.
type EncodeFunc[S, T any] func(source S, target T) error

var _ Encoder[any, any] = EncodeFunc[any, any](func(source, target any) error { return nil })

// Encode calls the underlying function to perform encoding.
func (e EncodeFunc[S, T]) Encode(source S, target T) error {
	return e(source, target)
}
