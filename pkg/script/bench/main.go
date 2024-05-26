package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/siyul-park/uniflow/pkg/script/compiler"
	"github.com/siyul-park/uniflow/pkg/script/eval"
	"github.com/siyul-park/uniflow/pkg/script/lexer"
	"github.com/siyul-park/uniflow/pkg/script/object"
	"github.com/siyul-park/uniflow/pkg/script/parser"
	"github.com/siyul-park/uniflow/pkg/script/vm"
)

const inputTmpl = `
let fib = fn(x) {
	if (x == 0) {
		0
	} else {
		if (x == 1) {
			1
		} else {
			fib(x - 1) + fib(x - 2)
		}
	}
};
fib(%v)
`

func main() {
	engine := flag.String("engine", "vm", "use 'vm' or 'eval'")
	flag.Usage = usage
	flag.Parse()

	num, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		flag.Usage()
		os.Exit(2)
	}

	var (
		duration time.Duration
		result   object.Object
	)

	input := fmt.Sprintf(inputTmpl, num)
	program := parser.New(lexer.New(input)).ParseProgram()

	if *engine == "vm" {
		start := time.Now()

		c := compiler.New()
		if err := c.Compile(program); err != nil {
			fmt.Printf("compiler error: %s", err)
			return
		}

		machine := vm.New(c.Bytecode())

		if err := machine.Run(); err != nil {
			fmt.Printf("vm error: %s", err)
			return
		}

		duration = time.Since(start)
		result = machine.LastPoppedStackElem()
	} else {
		env := object.NewEnvironment()
		start := time.Now()
		result = eval.Eval(program, env)
		duration = time.Since(start)
	}

	fmt.Printf("engine=%s, result=%s, duration=%s\n", *engine, result.Inspect(), duration)
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <integer>\n\n", os.Args[0])
	flag.PrintDefaults()
}
