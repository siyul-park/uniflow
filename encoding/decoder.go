package encoding

// Decoder is an interface for decoding data.
type Decoder[S, T any] interface {
	// Decode decodes data from the source to the target.
	Decode(source S, target T) error
}

type decoder[S, T any] struct {
	decode func(source S, target T) error
}

var _ Decoder[any, any] = (*decoder[any, any])(nil)

// DecodeFunc returns a Decoder implemented via a function.
func DecodeFunc[S, T any](decode func(source S, target T) error) Decoder[S, T] {
	return &decoder[S, T]{decode: decode}
}

// Decode calls the underlying decode function.
func (d *decoder[S, T]) Decode(source S, target T) error {
	return d.decode(source, target)
}
