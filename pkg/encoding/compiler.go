package encoding

import (
	"reflect"
	"unsafe"
)

// Compiler represents an interface for compiling encoders.
type Compiler[S any] interface {
	Compile(typ reflect.Type) (Encoder[S, unsafe.Pointer], error)
}

// CompilerFunc is a function type that implements the Compiler interface.
type CompilerFunc[S any] func(typ reflect.Type) (Encoder[S, unsafe.Pointer], error)

var _ Compiler[any] = CompilerFunc[any](func(typ reflect.Type) (Encoder[any, unsafe.Pointer], error) { return nil, nil })

func (c CompilerFunc[S]) Compile(typ reflect.Type) (Encoder[S, unsafe.Pointer], error) {
	return c(typ)
}
