package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/siyul-park/uniflow/pkg/script/compiler"
	"github.com/siyul-park/uniflow/pkg/script/eval"
	"github.com/siyul-park/uniflow/pkg/script/lexer"
	"github.com/siyul-park/uniflow/pkg/script/object"
	"github.com/siyul-park/uniflow/pkg/script/parser"
	"github.com/siyul-park/uniflow/pkg/script/vm"
)

const prompt = ">> "

// Start starts Monkey REPL.
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	macroEnv := object.NewEnvironment()

	symbolTable := compiler.NewSymbolTable()

	// Define built-in functions
	for i, builtin := range object.Builtins {
		symbolTable.DefineBuiltin(i, builtin.Name)
	}

	constants := make([]object.Object, 0)
	globals := make([]object.Object, vm.GlobalSize)

	for {
		fmt.Print(prompt)
		if !scanner.Scan() {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		// Process macros
		eval.DefineMacros(program, macroEnv)
		expanded := eval.ExpandMacros(program, macroEnv)

		// Compile the AST to bytecode
		complr := compiler.NewWithState(symbolTable, constants)
		if err := complr.Compile(expanded); err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed: %s\n", err)
			continue
		}

		// Update constant pool
		code := complr.Bytecode()
		constants = code.Constants

		// Run bytecode instructions
		machine := vm.NewWithGlobalStore(code, globals)
		if err := machine.Run(); err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed: %s\n", err)
			continue
		}

		lastPopped := machine.LastPoppedStackElem()
		if lastPopped == nil {
			io.WriteString(out, "no object at top of stack\n")
			continue
		}

		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, msg)
		io.WriteString(out, "\n")
	}
}
