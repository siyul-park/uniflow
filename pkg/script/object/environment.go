package object

// Environment associates values with variable names.
type Environment interface {
	// Get retrieves the value of a variable named by the `name`.
	// If the variable is present in the environment the value is returned and the boolean is true.
	// Otherwise the returned value will be nil and the boolean will be false.
	Get(name string) (Object, bool)

	// Set sets the `val` of a variable named by the `name` and returns the `val` itself.
	Set(name string, val Object) Object
}

// environment implements Environment interface.
// environment is not thread safe, so do not use it in multiple goroutines.
type environment struct {
	store map[string]Object
	outer Environment
}

// NewEnvironment returns a new Environment.
func NewEnvironment() Environment {
	return &environment{
		store: make(map[string]Object),
		outer: nil,
	}
}

// Get retrieves the value of a variable named by the `name`.
// If the variable is present in the environment the value is returned and the boolean is true.
// Otherwise the returned value will be nil and the boolean will be false.
func (e *environment) Get(name string) (Object, bool) {
	obj, exists := e.store[name]
	if !exists && e.outer != nil {
		obj, exists = e.outer.Get(name)
	}
	return obj, exists
}

// Set sets the `val` of a variable named by the `name` and returns the `val` itself.
func (e *environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

// NewEnclosedEnvironment creates a new Environment which holds the given outer Environment.
func NewEnclosedEnvironment(outer Environment) Environment {
	return &environment{
		store: make(map[string]Object),
		outer: outer,
	}
}
