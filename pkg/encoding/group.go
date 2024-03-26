package encoding

import (
	"github.com/pkg/errors"
	"sync"
)

// EncoderGroup is a group of encoders.
type EncoderGroup[S, T any] struct {
	encoders []Encoder[S, T]
	mu       sync.RWMutex
}

var _ Encoder[any, any] = (*EncoderGroup[any, any])(nil)

// NewEncoderGroup creates a new EncoderGroup instance.
func NewEncoderGroup[S, T any]() *EncoderGroup[S, T] {
	return &EncoderGroup[S, T]{}
}

// Add adds an encoder to the group.
func (d *EncoderGroup[S, T]) Add(encoder Encoder[S, T]) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.encoders = append(d.encoders, encoder)
}

// Len returns the number of encoders in the group.
func (d *EncoderGroup[S, T]) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	return len(d.encoders)
}

// Encode attempts to encode the source using the encoders in the group.
func (d *EncoderGroup[S, T]) Encode(source S, target T) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var err error
	for _, enc := range d.encoders {
		if err = enc.Encode(source, target); err == nil {
			return nil
		} else if !errors.Is(err, ErrUnsupportedValue) {
			return err
		}
	}

	return err
}
