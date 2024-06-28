package encoding

import (
	"reflect"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
)

// EncodeAssembler compiles and executes encoders for a specific source type.
type EncodeAssembler[S, T any] struct {
	compilers []EncodeCompiler[S, T]
	encoders  sync.Map
	mu        sync.RWMutex
}

// DecodeAssembler compiles and executes decoders for a specific target type.
type DecodeAssembler[S, T any] struct {
	compilers []DecodeCompiler[S]
	decoders  sync.Map
	mu        sync.RWMutex
}

var _ EncodeCompiler[any, any] = (*EncodeAssembler[any, any])(nil)
var _ Encoder[any, any] = (*EncodeAssembler[any, any])(nil)

var _ DecodeCompiler[any] = (*DecodeAssembler[any, any])(nil)
var _ Decoder[any, any] = (*DecodeAssembler[any, any])(nil)

// NewEncodeAssembler creates a new EncodeAssembler instance.
func NewEncodeAssembler[S, T any]() *EncodeAssembler[S, T] {
	return &EncodeAssembler[S, T]{}
}

// Add adds a compiler to the EncodeAssembler.
func (a *EncodeAssembler[S, T]) Add(compiler EncodeCompiler[S, T]) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.compilers = append(a.compilers, compiler)
}

// Len returns the number of compilers in the EncodeAssembler.
func (a *EncodeAssembler[S, T]) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.compilers)
}

// Encode encodes the source into the target type.
func (a *EncodeAssembler[S, T]) Encode(source S) (T, error) {
	enc, err := a.Compile(reflect.TypeOf(source))
	if err != nil {
		var zero T
		return zero, nil
	}
	return enc.Encode(source)
}

// Compile compiles an encoder for a given type.
func (a *EncodeAssembler[S, T]) Compile(typ reflect.Type) (Encoder[S, T], error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if enc, ok := a.encoders.Load(typ); ok {
		return enc.(Encoder[S, T]), nil
	}

	encoderGroup := NewEncoderGroup[S, T]()
	a.encoders.Store(typ, encoderGroup)

	for _, compiler := range a.compilers {
		if enc, err := compiler.Compile(typ); err == nil {
			encoderGroup.Add(enc)
		}
	}

	if encoderGroup.Len() == 0 {
		a.encoders.Delete(typ)
		return nil, errors.WithStack(ErrUnsupportedValue)
	}

	return encoderGroup, nil
}

// NewDecodeAssembler creates a new DecodeAssembler instance.
func NewDecodeAssembler[S, T any]() *DecodeAssembler[S, T] {
	return &DecodeAssembler[S, T]{}
}

// Add adds a compiler to the DecodeAssembler.
func (a *DecodeAssembler[S, T]) Add(compiler DecodeCompiler[S]) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.compilers = append(a.compilers, compiler)
}

// Len returns the number of compilers in the DecodeAssembler.
func (a *DecodeAssembler[S, T]) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.compilers)
}

// Decode decodes the source into the target.
func (a *DecodeAssembler[S, T]) Decode(source S, target T) error {
	val := reflect.ValueOf(target)
	ptr := val.UnsafePointer()

	dec, err := a.Compile(val.Type())
	if err != nil {
		return err
	}
	return dec.Decode(source, ptr)
}

// Compile compiles a decoder for a given type.
func (a *DecodeAssembler[S, T]) Compile(typ reflect.Type) (Decoder[S, unsafe.Pointer], error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if dec, ok := a.decoders.Load(typ); ok {
		return dec.(Decoder[S, unsafe.Pointer]), nil
	}

	var decoders []Decoder[S, unsafe.Pointer]
	for _, compiler := range a.compilers {
		if dec, err := compiler.Compile(typ); err == nil {
			decoders = append(decoders, dec)
		}
	}
	if len(decoders) == 0 {
		return nil, errors.WithStack(ErrUnsupportedValue)
	}

	decoderGroup := NewDecoderGroup[S, unsafe.Pointer]()
	for _, dec := range decoders {
		decoderGroup.Add(dec)
	}

	a.decoders.Store(typ, decoderGroup)
	return decoderGroup, nil
}
