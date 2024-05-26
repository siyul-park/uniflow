package object

import (
	"fmt"
)

// Builtins is a list of built-in functions.
var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		Name: "len",
		Builtin: &Builtin{
			Fn: func(args ...Object) Object {
				if l := len(args); l != 1 {
					return newError("wrong number of arguments. want=1, got=%d", l)
				}

				switch arg := args[0].(type) {
				case *String:
					return &Integer{Value: int64(len(arg.Value))}
				case *Array:
					return &Integer{Value: int64(len(arg.Elements))}
				default:
					return newError("argument to `len` not supported, got %s", arg.Type())
				}
			},
		},
	},
	{
		Name: "puts",
		Builtin: &Builtin{
			Fn: func(args ...Object) Object {
				for _, arg := range args {
					fmt.Println(arg.Inspect())
				}
				return nil
			},
		},
	},
	{
		Name: "first",
		Builtin: &Builtin{
			Fn: func(args ...Object) Object {
				if l := len(args); l != 1 {
					return newError("wrong number of arguments. want=1, got=%d", l)
				}

				if typ := args[0].Type(); typ != ArrayType {
					return newError("argument to `first` must be Array, got %s", typ)
				}

				arr := args[0].(*Array)
				if len(arr.Elements) > 0 {
					return arr.Elements[0]
				}
				return nil
			},
		},
	},
	{
		Name: "last",
		Builtin: &Builtin{
			Fn: func(args ...Object) Object {
				if l := len(args); l != 1 {
					return newError("wrong number of arguments. want=1, got=%d", l)
				}

				if typ := args[0].Type(); typ != ArrayType {
					return newError("argument to `last` must be Array, got %s", typ)
				}

				arr := args[0].(*Array)
				if l := len(arr.Elements); l > 0 {
					return arr.Elements[l-1]
				}
				return nil
			},
		},
	},
	{
		Name: "rest",
		Builtin: &Builtin{
			Fn: func(args ...Object) Object {
				if l := len(args); l != 1 {
					return newError("wrong number of arguments. want=1, got=%d", l)
				}

				if typ := args[0].Type(); typ != ArrayType {
					return newError("argument to `last` must be Array, got %s", typ)
				}

				arr := args[0].(*Array)
				l := len(arr.Elements)
				if l == 0 {
					return nil
				}

				newElems := make([]Object, l-1)
				copy(newElems, arr.Elements[1:l])
				return &Array{Elements: newElems}
			},
		},
	},
	{
		Name: "push",
		Builtin: &Builtin{
			Fn: func(args ...Object) Object {
				if l := len(args); l != 2 {
					return newError("wrong number of arguments. want=%d, got=%d", 2, l)
				}

				if typ := args[0].Type(); typ != ArrayType {
					return newError("first argument to `push` must be Array, got %s", typ)
				}

				arr := args[0].(*Array)
				l := len(arr.Elements)

				newElems := make([]Object, l+1)
				copy(newElems, arr.Elements)
				newElems[l] = args[1]
				return &Array{Elements: newElems}
			},
		},
	},
}

// GetBuiltinByName returns a built-in function matching a given name.
// If no function is found with the name, it returns nil.
func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}

	return nil
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}
