package encoding

import (
	"errors"
	"sync"
)

// EncoderGroup manages a group of encoders.
type EncoderGroup[S, T any] struct {
	encoders []Encoder[S, T]
	mu       sync.RWMutex
}

// DecoderGroup manages a group of decoders.
type DecoderGroup[S, T any] struct {
	decoders []Decoder[S, T]
	mu       sync.RWMutex
}

var _ Encoder[any, any] = (*EncoderGroup[any, any])(nil)
var _ Decoder[any, any] = (*DecoderGroup[any, any])(nil)

// NewDecoderGroup creates a new DecoderGroup.
func NewDecoderGroup[S, T any]() *DecoderGroup[S, T] {
	return &DecoderGroup[S, T]{}
}

// NewEncoderGroup creates a new EncoderGroup.
func NewEncoderGroup[S, T any]() *EncoderGroup[S, T] {
	return &EncoderGroup[S, T]{}
}

// Add adds an encoder to the group.
func (g *EncoderGroup[S, T]) Add(encoder Encoder[S, T]) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.encoders = append(g.encoders, encoder)
}

// Len returns the number of encoders in the group.
func (g *EncoderGroup[S, T]) Len() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.encoders)
}

// Encode attempts to encode the source using the encoders in the group.
func (g *EncoderGroup[S, T]) Encode(source S) (T, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var target T
	var err error
	for _, enc := range g.encoders {
		if target, err = enc.Encode(source); err == nil {
			return target, nil
		} else if !errors.Is(err, ErrUnsupportedValue) {
			return target, err
		}
	}
	return target, err
}

// Add adds a decoder to the group.
func (g *DecoderGroup[S, T]) Add(decoder Decoder[S, T]) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.decoders = append(g.decoders, decoder)
}

// Len returns the number of decoders in the group.
func (g *DecoderGroup[S, T]) Len() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.decoders)
}

// Decode attempts to decode the source using the decoders in the group.
func (g *DecoderGroup[S, T]) Decode(source S, target T) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var err error
	for _, dec := range g.decoders {
		if err = dec.Decode(source, target); err == nil {
			return nil
		} else if !errors.Is(err, ErrUnsupportedValue) {
			return err
		}
	}
	return err
}
