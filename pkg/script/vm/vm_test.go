package vm

import (
	"fmt"
	"testing"

	"github.com/siyul-park/uniflow/pkg/script/ast"
	"github.com/siyul-park/uniflow/pkg/script/compiler"
	"github.com/siyul-park/uniflow/pkg/script/lexer"
	"github.com/siyul-park/uniflow/pkg/script/object"
	"github.com/siyul-park/uniflow/pkg/script/parser"
)

type vmTestCase struct {
	input string
	want  interface{}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"50 * 2 * 2 + 10 - 5", 205},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 - 3) * 2 + -10", 64},
	}

	runVMTests(t, tests)
}

func TestFloatArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1.0", 1.0},
		{"1.1", 1.1},
		{"2.2", 2.2},
		{"1.25 + 2.25", 3.5},
		{"1.5 - 2.25", -0.75},
		{"1.25 * 2.5", 3.125},
		{"4.4 / 2.2", 2.0},
		{"3.3 / 1.2", 2.75},
		{"4 / 2", 2.0},
		{"3 / 2", 1.5},
		{"3.0 / 2.0", 1.5},
		{"3.5 / 2", 1.75},
		{"3 / 2.5", 1.2},
		{"50 / 2 * 2 + 10 - 5", 55.0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50.0},
		{"50 / 2 * 2.5 + 10.25 - 5.5", 67.25},
		{"5.125 + 5.250 + 5.375 + 5.500 - 10.625", 10.625},
		{"2 * 2 * 2 * 2 * 2.5", 40.0},
		{"5.5 * 2.2 + 10", 22.1},
		{"5.555 + 2.222 * 10.0", 27.775},
		{"5.5 * (2.2 + 10.1)", 67.65},
		{"-5.5", -5.5},
		{"-10.0", -10.0},
		{"-50.5 + 101 + -50.5", 0.0},
		{"(5.5 + 10 * 2.2 + 15 / 3) * 2.2 + -10", 61.5},
	}

	runVMTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 <= 1", true},
		{"1 >= 1", true},
		{"2 <= 1", false},
		{"1 >= 2", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"1.1 < 2.2", true},
		{"1.1 > 2.2", false},
		{"1.1 < 1.1", false},
		{"1.1 > 1.1", false},
		{"1.1 <= 1.1", true},
		{"1.1 >= 1.1", true},
		{"2.2 <= 1.1", false},
		{"1.1 >= 2.2", false},
		{"1.1 == 1.1", true},
		{"1.1 != 1.1", false},
		{"1.1 == 2.2", false},
		{"1.1 != 2.2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"true && false", false},
		{"false || true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"(1 <= 2) == true", true},
		{"(1 <= 2) == false", false},
		{"(1 >= 2) == true", false},
		{"(1 >= 2) == false", true},
		{"(1.1 < 2.2) == true", true},
		{"(1.1 < 2.2) == false", false},
		{"(1.1 > 2.2) == true", false},
		{"(1.1 > 2.2) == false", true},
		{"(1.1 <= 2.2) == true", true},
		{"(1.1 <= 2.2) == false", false},
		{"(1.1 >= 2.2) == true", false},
		{"(1.1 >= 2.2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!5.5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{"!!5.5", true},
		{"!(if (false) { 5 })", true},
		{"!(if (false) { 5.5 })", true},
		{"if ((if (false) { 10 })) { 10 } else { 20 }", 20},
		{"if ((if (false) { 10.5 })) { 10 } else { 20.5 }", 20.5},
	}

	runVMTests(t, tests)
}

func TestNilExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"nil", &object.Nil{}},
	}

	runVMTests(t, tests)
}

func TestLogicalExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"1 && 2", 2},
		{`"a" && 1`, 1},
		{"false && 1", false},
		{"1 || 2", 1},
		{`"a" || 1`, "a"},
		{"true || 1", true},
		{"1.1 && 2.2", 2.2},
		{"false && 1.1", false},
		{"1.1 || 2.2", 1.1},
		{"true || 1", true},
		{"true && false", false},
		{"false || true", true},
	}

	runVMTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 }", 20},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", Nil},
		{"if (1 <= 2) { 10 }", 10},
		{"if (1 <= 2) { 10 } else { 20 }", 10},
		{"if (1 >= 2) { 10 } else { 20 }", 20},
		{"if (1 >= 2) { 10 }", Nil},
		{"if (false) { 10 }", Nil},
	}

	runVMTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one = 1; one", 1},
		{"let one = 1; let two = 2; one + two", 3},
		{"let one = 1; let two = one + one; one + two", 3},
	}

	runVMTests(t, tests)
}

func TestGlobalAssignmentStatements(t *testing.T) {
	tests := []vmTestCase{
		{`one = 1; one`, 1},
		{`one = 1; two = one; two;`, 1},
		{`a = 1; a = 2; a`, 2},
		{`a = 1; b = 2; tmp = a; a = b; b = tmp; b`, 1},
	}

	runVMTests(t, tests)
}

func TestAssignmentStatementScopes(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			num = 55;
			fn() { num }();
			`,
			want: 55,
		},
		{
			input: `
			fn() {
				num = 55;
				num
			}();
			`,
			want: 55,
		},
		{
			input: `
			fn() {
				a = 55;
				b = 77;
				a + b;
			}();
			`,
			want: 132,
		},
		{
			input: `
			num = 55;
			fn() { num = 66 }();
			num;
			`,
			want: 55,
		},
		{
			input: `
			fn() {
				num = 55;
				num = 66;
				num;
			}();
			`,
			want: 66,
		},
		{
			input: `
			fn() {
				a = 55;
				b = 66;
				fn() {
					a = 77;
					b = 88;
					a + b;
				}();
				a + b;
			}();
			`,
			want: 121,
		},
	}

	runVMTests(t, tests)
}

func TestShadowingBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let len = 1;
			len;
			`,
			want: 1,
		},
		{
			input: `
			len = 1;
			len;
			`,
			want: 1,
		},
		{
			input: `
			len = 1;
			push = fn() {
				len = 2;
				len;
			}();
			len + push;
			`,
			want: 3,
		},
	}

	runVMTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"monkey"`, "monkey"},
		{`"mon" + "key"`, "monkey"},
		{`"mon" + "key" + "banana"`, "monkeybanana"},
	}

	runVMTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 - 4, 5 * 6]", []int{3, -1, 30}},
	}

	runVMTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			input: "{}",
			want:  map[object.HashKey]int64{},
		},
		{
			input: "{1: 2, 2: 3}",
			want: map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 2}).HashKey(): 3,
			},
		},
		{
			input: "{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			want: map[object.HashKey]int64{
				(&object.Integer{Value: 2}).HashKey(): 4,
				(&object.Integer{Value: 6}).HashKey(): 16,
			},
		},
	}

	runVMTests(t, tests)
}

func TestSetIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"a = [1, 2, 3]; a[1] = 4; a", []int{1, 4, 3}},
		{"a = [1, 2, 3]; a[0 + 2] = 3 - 4; a", []int{1, 2, -1}},
		{"a = [[1, 1, 1]]; a[0][0] = 2; a[0]", []int{2, 1, 1}},
		{
			input: "h = {}; h[0] = 1; h",
			want: map[object.HashKey]int64{
				(&object.Integer{Value: 0}).HashKey(): 1,
			},
		},
		{
			input: "h = {1: 1, 2: 2}; h[1] = 0; h",
			want: map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 0,
				(&object.Integer{Value: 2}).HashKey(): 2,
			},
		},
		{
			input: "h = {1: 1, 2: 2}; h[3] = 3; h",
			want: map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 1,
				(&object.Integer{Value: 2}).HashKey(): 2,
				(&object.Integer{Value: 3}).HashKey(): 3,
			},
		},
	}

	runVMTests(t, tests)
}

func TestSetIndexExpressionErrors(t *testing.T) {
	tests := []string{
		"a = []; a[1] = 1",
		"a = [1, 2, 3]; a[10] = 9",
		"a = [[1, 1, 1]]; a[1][0] = 2",
	}

	runVMTestErrors(t, tests)
}

func TestGetIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", Nil},
		{"[1, 2, 3][99]", Nil},
		{"[1][-1]", Nil},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[0]", Nil},
		{"{}[0]", Nil},
	}

	runVMTests(t, tests)
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let fivePlusTen = fn() { 5 + 10; };
			fivePlusTen();
			`,
			want: 15,
		},
		{
			input: `
			let one = fn() { 1; };
			let two = fn() { 2; };
			one() + two();
			`,
			want: 3,
		},
		{
			input: `
			let a = fn() { 1; };
			let b = fn() { a() + 1; };
			let c = fn() { b() + 1; };
			c();
			`,
			want: 3,
		},
	}

	runVMTests(t, tests)
}

func TestFunctionsWithReturnStatements(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let earlyExit = fn() { return 99; 100; };
			earlyExit();
			`,
			want: 99,
		},
		{
			input: `
			let earlyExit = fn() { return 99; return 100; };
			earlyExit();
			`,
			want: 99,
		},
	}

	runVMTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let noReturn = fn() { };
			noReturn();
			`,
			want: Nil,
		},
		{
			input: `
			let noReturn = fn() { };
			let noReturnTwo = fn() { noReturn(); };
			noReturn();
			noReturnTwo();
			`,
			want: Nil,
		},
	}

	runVMTests(t, tests)
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let returnsOne = fn() { 1; };
			let returnsOneReturner = fn() { returnsOne; };
			returnsOneReturner()();
			`,
			want: 1,
		},
		{
			input: `
			let returnsOneReturner = fn() {
				let returnsOne = fn() { 1; };
				returnsOne;
			};
			returnsOneReturner()();
			`,
			want: 1,
		},
	}

	runVMTests(t, tests)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let one = fn() { let one = 1; one };
			one();
			`,
			want: 1,
		},
		{
			input: `
			let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
			oneAndTwo();
			`,
			want: 3,
		},
		{
			input: `
			let oneAndTwo = fn() { let one = 1; let two = 2; one + two };
			let threeAndFour = fn() { let three = 3; let four = 4; three + four; };
			oneAndTwo() + threeAndFour();
			`,
			want: 10,
		},
		{
			input: `
			let firstFoobar = fn() { let foobar = 50; foobar; };
			let secondFoobar = fn() { let foobar = 100; foobar; };
			firstFoobar() + secondFoobar();
			`,
			want: 150,
		},
		{
			input: `
			let globalSeed = 50;
			let minusOne = fn() {
				let num = 1;
				globalSeed - num;
			};
			let minusTwo = fn() {
				let num = 2;
				globalSeed - num;
			};
			minusOne() + minusTwo();
			`,
			want: 97,
		},
	}

	runVMTests(t, tests)
}

func TestCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let id = fn(a) { a; };
			id(4);
			`,
			want: 4,
		},
		{
			input: `
			let sum = fn(a, b) { a + b; };
			sum(1, 2);
			`,
			want: 3,
		},
		{
			input: `
			let sum = fn(a, b) {
				let c = a + b;
				c;
			};
			sum(1, 2);
			`,
			want: 3,
		},
		{
			input: `
			let sum = fn(a, b) {
				let c = a + b;
				c;
			};
			sum(1, 2) + sum(3, 4);
			`,
			want: 10,
		},
		{
			input: `
			let sum = fn(a, b) {
				let c = a + b;
				c;
			};
			let outer = fn() {
				sum(1, 2) + sum(3, 4);
			};
			outer();
			`,
			want: 10,
		},
		{
			input: `
			let globalNum = 10;

			let sum = fn(a, b) {
				let c = a + b;
				c + globalNum;
			};

			let outer = fn() {
				sum(1, 2) + sum(3, 4) + globalNum;
			};

			outer() + globalNum;
			`,
			want: 50,
		},
	}

	runVMTests(t, tests)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: "fn() { 1; }(1);",
			want:  "wrong number of arguments: want=0, got=1",
		},
		{
			input: "fn(a) { a; }();",
			want:  "wrong number of arguments: want=1, got=0",
		},
		{
			input: "fn(a, b) { a + b; }(1);",
			want:  "wrong number of arguments: want=2, got=1",
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)

		c := compiler.New()
		if err := c.Compile(program); err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(c.Bytecode())
		if err := vm.Run(); err == nil {
			t.Fatalf("expected VM error but resulted in none")
		} else if err.Error() != tt.want {
			t.Fatalf("wrong VM error: want=%q, got=%q", tt.want, err)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, &object.Error{Message: "argument to `len` not supported, got Integer"}},
		{`len("one", "two")`, &object.Error{Message: "wrong number of arguments. want=1, got=2"}},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`puts("hello", "world!")`, Nil},
		{`first([1, 2, 3])`, 1},
		{`first([])`, Nil},
		{`first(1)`, &object.Error{Message: "argument to `first` must be Array, got Integer"}},
		{`rest([1, 2, 3])`, []int{2, 3}},
		{`rest([])`, Nil},
		{`push([], 1)`, []int{1}},
		{`push(1, 1)`, &object.Error{Message: "first argument to `push` must be Array, got Integer"}},
		{`first(rest(push([1, 2, 3], 4)))`, 2},
	}

	runVMTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let newClosure = fn(a) {
				fn() { a; };
			};
			let closure = newClosure(99);
			closure();
			`,
			want: 99,
		},
		{
			input: `
			let newAdder = fn(a, b) {
				fn(c) { a + b + c };
			};
			let adder = newAdder(1, 2);
			adder(8);
			`,
			want: 11,
		},
		{
			input: `
			let newAdder = fn(a, b) {
				let c = a + b;
				fn(d) { c + d };
			};
			let adder = newAdder(1, 2);
			adder(8);
			`,
			want: 11,
		},
		{
			input: `
			let newAdderOuter = fn(a, b) {
				let c = a + b;
				fn(d) {
					let e = c + d;
					fn(f) { e + f; };
				};
			};
			let newAdderInner = newAdderOuter(1, 2);
			let adder = newAdderInner(3);
			adder(8);
			`,
			want: 14,
		},
		{
			input: `
			let a = 1;
			let newAdderOuter = fn(b) {
				fn(c) {
					fn(d) { a + b + c + d; };
				};
			};
			let newAdderInner = newAdderOuter(2);
			let adder = newAdderInner(3);
			adder(8);
			`,
			want: 14,
		},
		{
			input: `
			let newClosure = fn(a, b) {
				let one = fn() { a; };
				let two = fn() { b; };
				fn() { one() + two(); };
			};
			let closure = newClosure(9, 90);
			closure();
			`,
			want: 99,
		},
	}

	runVMTests(t, tests)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let countDown = fn(x) {
				if (x == 0) {
					return 0;
				} else {
					countDown(x - 1);
				}
			};
			countDown(1);
			`,
			want: 0,
		},
		{
			input: `
			let countDown = fn(x) {
				if (x == 0) {
					return 0;
				} else {
					countDown(x - 1);
				}
			};
			let wrapper = fn() { countDown(1); };
			wrapper();
			`,
			want: 0,
		},
		{
			input: `
			let wrapper = fn() {
				let countDown = fn(x) {
					if (x == 0) {
						return 0;
					} else {
						countDown(x - 1);
					}
				};
				countDown(1);
			};
			wrapper();
			`,
			want: 0,
		},
	}

	runVMTests(t, tests)
}

func TestRecursiveFibonacci(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let fib = fn(x) {
				if (x == 0) {
					return 0;
				} else {
					if (x == 1) {
						return 1;
					} else {
						fib(x - 1) + fib(x - 2);
					}
				}
			};
			fib(15);
			`,
			want: 610,
		},
	}

	runVMTests(t, tests)
}

func runVMTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		complr := compiler.New()
		if err := complr.Compile(program); err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		// dumpBytecode(complr.Bytecode())

		vm := New(complr.Bytecode())
		if err := vm.Run(); err != nil {
			t.Fatalf("vm error: %s", err)
		}

		got := vm.LastPoppedStackElem()

		testExpectedObject(t, tt.want, got)
	}
}

func runVMTestErrors(t *testing.T, tests []string) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt)

		complr := compiler.New()
		if err := complr.Compile(program); err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		// dumpBytecode(complr.Bytecode())

		vm := New(complr.Bytecode())
		if err := vm.Run(); err == nil {
			t.Errorf("expected vm error, but got nil")
		}
	}
}

func dumpBytecode(bytecode *compiler.Bytecode) {
	fmt.Println("===== Instructions =====")

	fmt.Println(bytecode.Instructions)

	fmt.Println("===== Constants =====")

	for i, c := range bytecode.Constants {
		fmt.Printf("Constant %d %p (%T):\n", i, c, c)

		switch c := c.(type) {
		case *object.CompiledFunction:
			fmt.Printf("  Instructions:\n%s", c.Instructions)
		case *object.Integer:
			fmt.Printf("  Value: %d\n", c.Value)
		}
	}

	fmt.Printf("\n")
}

func parse(input string) *ast.Program {
	return parser.New(lexer.New(input)).ParseProgram()
}

func testExpectedObject(t *testing.T, want interface{}, got object.Object) {
	t.Helper()

	switch want := want.(type) {
	case bool:
		if err := testBooleanObject(bool(want), got); err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}

	case int:
		if err := testIntegerObject(int64(want), got); err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}

	case float64:
		if err := testFloatObject(want, got); err != nil {
			t.Errorf("testFloatObject failed: %s", err)
		}

	case string:
		if err := testStringObject(want, got); err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}

	case []int:
		arr, ok := got.(*object.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%#v)", got, got)
			return
		}

		if len(arr.Elements) != len(want) {
			t.Errorf("wrong num of elements. want=%d, got=%d", len(want), len(arr.Elements))
			return
		}

		for i, el := range want {
			if err := testIntegerObject(int64(el), arr.Elements[i]); err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}

	case map[object.HashKey]int64:
		hash, ok := got.(*object.Hash)
		if !ok {
			t.Errorf("object is not Hash. got=%T (%#v)", got, got)
		}

		if len(hash.Pairs) != len(want) {
			t.Errorf(
				"hash has wrong number of pairs. want=%d (%#v), got=%d (%#v)",
				len(want), want, len(hash.Pairs), hash.Pairs,
			)
		}

		for wantKey, wantVal := range want {
			pair, ok := hash.Pairs[wantKey]
			if !ok {
				t.Errorf("no pair for given key %v in pairs", wantKey)
			}

			if err := testIntegerObject(wantVal, pair.Value); err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}

	case *object.Nil:
		if got != Nil {
			t.Errorf("object is not Nil: %T (%#v)", got, got)
		}

	case *object.Error:
		err, ok := got.(*object.Error)
		if !ok {
			t.Errorf("object is not Error: %T (%+v)", got, got)
		}

		if err.Message != want.Message {
			t.Errorf("wrong error message. want=%q, got=%q", want.Message, err.Message)
		}

	default:
		t.Errorf("testExpectedObject failed: unknown type %T (%#v)", got, got)
	}
}

func testBooleanObject(want bool, got object.Object) error {
	result, ok := got.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%#v)", got, got)
	}

	if result.Value != want {
		return fmt.Errorf("object has wrong value. want=%t, got=%t", want, result.Value)
	}

	return nil
}

func testIntegerObject(want int64, got object.Object) error {
	result, ok := got.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%#v)", got, got)
	}

	if result.Value != want {
		return fmt.Errorf("object has wrong value. want=%d, got=%d", want, result.Value)
	}

	return nil
}

func testFloatObject(want float64, got object.Object) error {
	result, ok := got.(*object.Float)
	if !ok {
		return fmt.Errorf("object is not Float. got=%T (%#v)", got, got)
	}

	if result.Value != want {
		return fmt.Errorf("object has wrong value. want=%v, got=%v", want, result.Value)
	}

	return nil
}

func testStringObject(want string, got object.Object) error {
	result, ok := got.(*object.String)
	if !ok {
		return fmt.Errorf("object is not String. got=%T (%#v)", got, got)

	}

	if result.Value != want {
		return fmt.Errorf("object has wrong value. want=%q, got=%q", want, result.Value)
	}

	return nil
}
