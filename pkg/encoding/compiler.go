package encoding

import (
	"github.com/pkg/errors"
	"reflect"
	"sync"
	"unsafe"
)

type CompiledDecoder[S, T any] struct {
	compilers []Compiler[S]
	decoders  map[reflect.Type]Decoder[S, unsafe.Pointer]
	mu        sync.RWMutex
}

type Compiler[S any] interface {
	Compile(typ reflect.Type) (Decoder[S, unsafe.Pointer], error)
}

type CompilerFunc[S any] func(typ reflect.Type) (Decoder[S, unsafe.Pointer], error)

var _ Compiler[any] = CompilerFunc[any](func(typ reflect.Type) (Decoder[any, unsafe.Pointer], error) { return nil, nil })

func (c CompilerFunc[S]) Compile(typ reflect.Type) (Decoder[S, unsafe.Pointer], error) {
	return c(typ)
}

var _ Compiler[any] = (*CompiledDecoder[any, any])(nil)
var _ Decoder[any, any] = (*CompiledDecoder[any, any])(nil)

func NewCompiledDecoder[S, T any]() *CompiledDecoder[S, T] {
	return &CompiledDecoder[S, T]{
		decoders: map[reflect.Type]Decoder[S, unsafe.Pointer]{},
	}
}

func (c *CompiledDecoder[S, T]) Add(compiler Compiler[S]) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.compilers = append(c.compilers, compiler)
}

func (c *CompiledDecoder[S, T]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.compilers)
}

func (c *CompiledDecoder[S, T]) Decode(source S, target T) error {
	typ := reflect.TypeOf(target)
	val := reflect.ValueOf(target)

	dec, err := c.Compile(typ)
	if err != nil {
		return err
	}

	return dec.Decode(source, val.UnsafePointer())
}

func (c *CompiledDecoder[S, T]) Compile(typ reflect.Type) (Decoder[S, unsafe.Pointer], error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if dec, ok := c.decoders[typ]; ok {
		return dec, nil
	}

	var decoders []Decoder[S, unsafe.Pointer]
	for _, compiler := range c.compilers {
		if dec, err := compiler.Compile(typ); err == nil {
			decoders = append(decoders, dec)
		}
	}
	if len(decoders) == 0 {
		return nil, errors.WithStack(ErrUnsupportedValue)
	}

	decoder := NewDecoderGroup[S, unsafe.Pointer]()
	for _, d := range decoders {
		decoder.Add(d)
	}
	c.decoders[typ] = decoder
	return decoder, nil
}
