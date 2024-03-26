package encoding

import (
	"github.com/pkg/errors"
	"reflect"
	"sync"
	"unsafe"
)

// Assembler compiles and executes encoders for a specific target type.
type Assembler[S, T any] struct {
	compilers []Compiler[S]
	encoders  sync.Map
	mu        sync.RWMutex
}

// Compiler represents an interface for compiling encoders.
type Compiler[S any] interface {
	Compile(typ reflect.Type) (Encoder[S, unsafe.Pointer], error)
}

// CompilerFunc is a function type that implements the Compiler interface.
type CompilerFunc[S any] func(typ reflect.Type) (Encoder[S, unsafe.Pointer], error)

var (
	_ Compiler[any]     = (*Assembler[any, any])(nil)
	_ Encoder[any, any] = (*Assembler[any, any])(nil)
	_ Compiler[any]     = CompilerFunc[any](func(typ reflect.Type) (Encoder[any, unsafe.Pointer], error) { return nil, nil })
)

// NewAssembler creates a new Assembler instance.
func NewAssembler[S, T any]() *Assembler[S, T] {
	return &Assembler[S, T]{}
}

// Add adds a compiler to the Assembler.
func (a *Assembler[S, T]) Add(compiler Compiler[S]) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.compilers = append(a.compilers, compiler)
}

// Len returns the number of compilers in the Assembler.
func (a *Assembler[S, T]) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.compilers)
}

// Encode encodes the source into the target.
func (a *Assembler[S, T]) Encode(source S, target T) error {
	typ := reflect.TypeOf(target)
	if typ == nil {
		return nil
	}

	enc, err := a.Compile(typ)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(target)

	var ptr unsafe.Pointer
	if typ.Kind() == reflect.Ptr {
		ptr = val.UnsafePointer()
	} else {
		zero := reflect.New(typ)
		zero.Elem().Set(val)
		ptr = zero.UnsafePointer()
	}

	return enc.Encode(source, ptr)
}

// Compile compiles an encoder for a given type.
func (a *Assembler[S, T]) Compile(typ reflect.Type) (Encoder[S, unsafe.Pointer], error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if typ == nil {
		return EncodeFunc[S, unsafe.Pointer](func(source S, target unsafe.Pointer) error {
			return nil
		}), nil
	}

	if typ.Kind() != reflect.Ptr {
		typ = reflect.PtrTo(typ)
	}

	if enc, ok := a.encoders.Load(typ); ok {
		return enc.(Encoder[S, unsafe.Pointer]), nil
	}

	var encoders []Encoder[S, unsafe.Pointer]
	for _, compiler := range a.compilers {
		if enc, err := compiler.Compile(typ); err == nil {
			encoders = append(encoders, enc)
		}
	}
	if len(encoders) == 0 {
		return nil, errors.WithStack(ErrUnsupportedValue)
	}

	encoder := NewEncoderGroup[S, unsafe.Pointer]()
	for _, enc := range encoders {
		encoder.Add(enc)
	}

	a.encoders.Store(typ, encoder)
	return encoder, nil
}

func (c CompilerFunc[S]) Compile(typ reflect.Type) (Encoder[S, unsafe.Pointer], error) {
	return c(typ)
}
