package encoding

import (
	"sync"

	"github.com/pkg/errors"
)

// EncoderGroup is a group of encoders.
type EncoderGroup[S, T any] struct {
	encoders []Encoder[S, T]
	lock     sync.RWMutex
}

var _ Encoder[any, any] = (*EncoderGroup[any, any])(nil)

// NewEncoderGroup creates a new EncoderGroup instance.
func NewEncoderGroup[S, T any]() *EncoderGroup[S, T] {
	return &EncoderGroup[S, T]{}
}

// Add adds an encoder to the group.
func (e *EncoderGroup[S, T]) Add(encoder Encoder[S, T]) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.encoders = append(e.encoders, encoder)
}

// Len returns the number of encoders in the group.
func (e *EncoderGroup[S, T]) Len() int {
	e.lock.Lock()
	defer e.lock.Unlock()

	return len(e.encoders)
}

// Encode attempts to encode the source using the encoders in the group.
func (e *EncoderGroup[S, T]) Encode(source S) (T, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	var zero T

	var target T
	var err error
	for _, encoder := range e.encoders {
		if target, err = encoder.Encode(source); err == nil {
			return target, nil
		} else if !errors.Is(err, ErrUnsupportedValue) {
			return zero, err
		}
	}

	return zero, err
}

// DecoderGroup is a group of decoders.
type DecoderGroup[S, T any] struct {
	decoders []Decoder[S, T]
	lock     sync.RWMutex
}

var _ Decoder[any, any] = (*DecoderGroup[any, any])(nil)

// NewDecoderGroup creates a new DecoderGroup instance.
func NewDecoderGroup[S, T any]() *DecoderGroup[S, T] {
	return &DecoderGroup[S, T]{}
}

// Add adds a decoder to the group.
func (d *DecoderGroup[S, T]) Add(decoder Decoder[S, T]) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.decoders = append(d.decoders, decoder)
}

// Len returns the number of decoders in the group.
func (d *DecoderGroup[S, T]) Len() int {
	d.lock.Lock()
	defer d.lock.Unlock()

	return len(d.decoders)
}

// Decode attempts to decode the source using the decoders in the group.
func (d *DecoderGroup[S, T]) Decode(source S, target T) error {
	d.lock.RLock()
	defer d.lock.RUnlock()

	var err error
	for _, decoder := range d.decoders {
		if err = decoder.Decode(source, target); err == nil {
			return nil
		} else if !errors.Is(err, ErrUnsupportedValue) {
			return err
		}
	}

	return err
}
