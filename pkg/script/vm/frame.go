package vm

import (
	"github.com/siyul-park/uniflow/pkg/script/code"
	"github.com/siyul-park/uniflow/pkg/script/object"
)

// Frame represents a stack frame.
type Frame struct {
	cl *object.Closure
	// Instruction pointer.
	ip int
	// Base pointer points to the bottom of the stack of the current stack frame.
	// It's also called "frame pointer".
	bp int
}

// NewFrame creates a new stack frame for a given compiled function.
func NewFrame(cl *object.Closure, bp int) *Frame {
	return &Frame{cl: cl, ip: -1, bp: bp}
}

// Instructions returns bytecode instructions of a function the stack frame is created for.
func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
