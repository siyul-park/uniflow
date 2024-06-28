package encoding

import (
	"reflect"
	"unsafe"
)

// EncodeCompiler compiles an encoder for a given type.
type EncodeCompiler[S, T any] interface {
	Compile(typ reflect.Type) (Encoder[S, T], error)
}

// DecodeCompiler compiles a decoder for a given type.
type DecodeCompiler[S any] interface {
	Compile(typ reflect.Type) (Decoder[S, unsafe.Pointer], error)
}

// EncodeCompilerFunc is a function type that implements EncodeCompiler.
type EncodeCompilerFunc[S, T any] func(typ reflect.Type) (Encoder[S, T], error)

// DecodeCompilerFunc is a function type that implements DecodeCompiler.
type DecodeCompilerFunc[S any] func(typ reflect.Type) (Decoder[S, unsafe.Pointer], error)

var _ EncodeCompiler[any, any] = EncodeCompilerFunc[any, any](nil)
var _ DecodeCompiler[any] = DecodeCompilerFunc[any](nil)

// Compile calls the underlying function to compile an encoder.
func (f EncodeCompilerFunc[S, T]) Compile(typ reflect.Type) (Encoder[S, T], error) {
	return f(typ)
}

// Compile calls the underlying function to compile a decoder.
func (f DecodeCompilerFunc[S]) Compile(typ reflect.Type) (Decoder[S, unsafe.Pointer], error) {
	return f(typ)
}
