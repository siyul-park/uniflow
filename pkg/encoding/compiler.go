package encoding

import (
	"github.com/pkg/errors"
	"reflect"
	"sync"
	"unsafe"
)

type CompiledDecoder[S, T any] struct {
	compilers []Compiler[S]
	decoders  sync.Map // map[reflect.Type]Decoder[S, unsafe.Pointer]
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
	return &CompiledDecoder[S, T]{}
}

func (c *CompiledDecoder[S, T]) Add(compiler Compiler[S]) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.compilers = append(c.compilers, compiler)
}

func (c *CompiledDecoder[S, T]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.compilers)
}

func (c *CompiledDecoder[S, T]) Decode(source S, target T) error {
	typ := reflect.TypeOf(target)
	if typ == nil {
		return nil
	}

	dec, err := c.Compile(typ)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(target)

	var ptr unsafe.Pointer
	if typ.Kind() == reflect.Pointer {
		ptr = val.UnsafePointer()
	} else {
		zero := reflect.New(typ)
		zero.Elem().Set(val)
		ptr = zero.UnsafePointer()
	}

	return dec.Decode(source, ptr)
}

func (c *CompiledDecoder[S, T]) Compile(typ reflect.Type) (Decoder[S, unsafe.Pointer], error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if typ == nil {
		return DecoderFunc[S, unsafe.Pointer](func(source S, target unsafe.Pointer) error {
			return nil
		}), nil
	}

	if typ.Kind() != reflect.Pointer {
		typ = reflect.PointerTo(typ)
	}

	if dec, ok := c.decoders.Load(typ); ok {
		return dec.(Decoder[S, unsafe.Pointer]), nil
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
	c.decoders.Store(typ, decoder)
	return decoder, nil
}
