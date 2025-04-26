package encoding

import (
	"errors"
	"reflect"
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
	cache    sync.Map
	mu       sync.RWMutex
}

var (
	ErrUnsupportedType  = errors.New("type is unsupported")
	ErrUnsupportedValue = errors.New("value is unsupported")
)

var (
	_ Encoder[any, any] = (*EncoderGroup[any, any])(nil)
	_ Decoder[any, any] = (*DecoderGroup[any, any])(nil)
)

// NewDecoderGroup creates a new DecoderGroup.
func NewDecoderGroup[S, T any]() *DecoderGroup[S, T] {
	return &DecoderGroup[S, T]{}
}

// NewEncoderGroup creates a new EncoderGroup.
func NewEncoderGroup[S, T any]() *EncoderGroup[S, T] {
	return &EncoderGroup[S, T]{}
}

// Add adds an encoder to the group.
func (g *EncoderGroup[S, T]) Add(encoder Encoder[S, T]) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, enc := range g.encoders {
		if enc == encoder {
			return false
		}
	}
	g.encoders = append(g.encoders, encoder)
	return true
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
		} else if !errors.Is(err, ErrUnsupportedType) {
			return target, err
		}
	}
	return target, err
}

// Add adds a decoder to the group.
func (g *DecoderGroup[S, T]) Add(decoder Decoder[S, T]) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, dec := range g.decoders {
		if dec == decoder {
			return false
		}
	}
	g.decoders = append(g.decoders, decoder)
	return true
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

	typ := reflect.TypeOf(source)

	var err error
	cache, ok := g.cache.Load(typ)
	if ok {
		if err = cache.(Decoder[S, T]).Decode(source, target); err == nil {
			return nil
		}
	}
	for _, dec := range g.decoders {
		if dec == cache {
			continue
		}
		if err = dec.Decode(source, target); err == nil {
			g.cache.Store(typ, dec)
			return nil
		} else if !errors.Is(err, ErrUnsupportedType) {
			return err
		}
	}
	return err
}
