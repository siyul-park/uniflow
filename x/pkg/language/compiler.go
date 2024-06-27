package language

// Compiler represents an interface for compiling source code into a Program.
type Compiler interface {
	Compile(string) (Program, error)
}

// CompileFunc is a function type that implements the Compiler interface.
type CompileFunc func(string) (Program, error)

var _ Compiler = CompileFunc(nil)

// Compile compiles the given source code into a Program using the CompileFunc.
func (f CompileFunc) Compile(code string) (Program, error) {
	return f(code)
}
