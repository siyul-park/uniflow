package language

// Compiler is an interface for compiling source code into a Program.
type Compiler interface {
	Compile(string) (Program, error)
}

type compiler struct {
	fn func(string) (Program, error)
}

// CompileFunc creates a new Compiler that uses the provided function to compile code.
func CompileFunc(fn func(string) (Program, error)) Compiler {
	return &compiler{fn: fn}
}

// Compile compiles the provided source code into a Program.
func (c *compiler) Compile(code string) (Program, error) {
	return c.fn(code)
}
