package encoding

import (
	"sync"

	"github.com/pkg/errors"
)

type (
	// EncoderGroup is a group of Encoder.
	EncoderGroup[S, T any] struct {
		encoders []Encoder[S, T]
		lock     sync.RWMutex
	}

	// DecoderGroup is a group of Decoder.
	DecoderGroup[S, T any] struct {
		decoders []Decoder[S, T]
		lock     sync.RWMutex
	}
)

var _ Encoder[any, any] = (*EncoderGroup[any, any])(nil)
var _ Decoder[any, any] = (*DecoderGroup[any, any])(nil)

func NewEncoderGroup[S, T any]() *EncoderGroup[S, T] {
	return &EncoderGroup[S, T]{}
}

func (e *EncoderGroup[S, T]) Add(encoder Encoder[S, T]) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.encoders = append(e.encoders, encoder)
}

func (e *EncoderGroup[S, T]) Len() int {
	e.lock.Lock()
	defer e.lock.Unlock()

	return len(e.encoders)
}

func (e *EncoderGroup[S, T]) Encode(source S) (T, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	var zero T
	for _, encoder := range e.encoders {
		if target, err := encoder.Encode(source); err == nil {
			return target, nil
		} else if !errors.Is(err, ErrUnsupportedValue) {
			return zero, err
		}
	}
	return zero, errors.WithStack(ErrUnsupportedValue)
}

func NewDecoderGroup[S, T any]() *DecoderGroup[S, T] {
	return &DecoderGroup[S, T]{}
}

func (d *DecoderGroup[S, T]) Add(decoder Decoder[S, T]) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.decoders = append(d.decoders, decoder)
}

func (d *DecoderGroup[S, T]) Len() int {
	d.lock.Lock()
	defer d.lock.Unlock()

	return len(d.decoders)
}

func (d *DecoderGroup[S, T]) Decode(source S, target T) error {
	d.lock.RLock()
	defer d.lock.RUnlock()

	for _, decoder := range d.decoders {
		if err := decoder.Decode(source, target); err == nil {
			return nil
		} else if !errors.Is(err, ErrUnsupportedValue) {
			return err
		}
	}
	return errors.WithStack(ErrUnsupportedValue)
}
