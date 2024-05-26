package eval

import (
	"strings"
	"testing"

	"github.com/siyul-park/uniflow/pkg/script/lexer"
	"github.com/siyul-park/uniflow/pkg/script/object"
	"github.com/siyul-park/uniflow/pkg/script/parser"
)

func testEval(t *testing.T, input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("input %q has errors: \n%v", input, strings.Join(p.Errors(), "\n"))
	}

	env := object.NewEnvironment()
	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Fatalf("object is not *object.Integer. got=%#v", obj)
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. want=%d, got=%d", expected, result.Value)
	}
}

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testFloatObject(t *testing.T, obj object.Object, expected float64) {
	f, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not *object.Float. got=%#v", obj)
		return
	}

	if f.Value != expected {
		t.Errorf("object has wrong value. want=%f, got=%f", expected, f.Value)
	}
}

func TestEvalFloatExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"12.34", 12.34},
		{"0.56", 0.56},
		{"78.00", 78.00},
		{"-12.34", -12.34},
		{"-0.56", -0.56},
		{"-78.00", -78.00},
		{"(5 + 10.0 * 2.5 + 15.0 / 3) * 2.1 + -10.1", 63.4},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Fatalf("object is not *object.Boolean. got=%#v", obj)
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. want=%t, got=%t", expected, result.Value)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{`"hello" == "hello"`, true},
		{`"hello" == "world"`, false},
		{`"foo" != "bar"`, true},
		{`"foo" != "foo"`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		if i, ok := tt.expected.(int); ok {
			testIntegerObject(t, evaluated, int64(i))
		} else {
			testNilObject(t, evaluated)
		}
	}
}

func testNilObject(t *testing.T, obj object.Object) {
	if obj != NilValue {
		t.Errorf("object is not NilValue. got=%#v", obj)
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 11;", 10},
		{`
		if (10 > 1) {
			if (10 > 1) {
				return 10;
			}

			return 1;
		}
		`, 10},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{"5 + true;", "type mismatch: Integer + Boolean"},
		{"5 + true; 5;", "type mismatch: Integer + Boolean"},
		{"-true", "unknown operator: -Boolean"},
		{"true + false", "unknown operator: Boolean + Boolean"},
		{"5; true + false; 5", "unknown operator: Boolean + Boolean"},
		{"if (10 > 1) { true + false; }", "unknown operator: Boolean + Boolean"},
		{`
		if (10 > 1) {
			if (10 > 1) {
				return true + false;
			}

			return 1;
		}
		`, "unknown operator: Boolean + Boolean"},
		{"foobar", "identifier not found: foobar"},
		{`"Hello" - "World"`, "unknown operator: String - String"},
		{`1.5 + "World"`, "unknown operator: Float + String"},
		{`{[1, 2]: "Monkey"}`, "unusable as hash key: Array"},
		{`{"name": "Monkey"}[fn(x) { x }]`, "unusable as hash key: Function"},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%#v", evaluated)
			continue
		}

		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expectedMessage, errObj.Message)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; }"

	evaluated := testEval(t, input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not *object.Function. got=%#v", evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. got=%+v", fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"
	if body := fn.Body.String(); body != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, body)
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5);", 5},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
	let newAdder = fn(x) {
		fn(y) { x + y };
	};

	let addTwo = newAdder(2);
	addTwo(2);
	`

	evaluated := testEval(t, input)
	testIntegerObject(t, evaluated, 4)
}

func TestStringLiteralAndConcat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"Hello World!";`, "Hello World!"},
		{`"Hello" + " " + "World!";`, "Hello World!"},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		str, ok := evaluated.(*object.String)
		if !ok {
			t.Fatalf("object is not *object.String. got=%#v", evaluated)
		}

		if str.Value != tt.expected {
			t.Errorf("String has wrong value. want=%q, got=%q", tt.expected, str.Value)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// len for strings
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len("hello" + " " + "world")`, 11},
		{`len(1)`, "argument to `len` not supported, got Integer"},
		{`len("one", "two")`, "wrong number of arguments. want=1, got=2"},
		// len for arrays
		{"len([])", 0},
		{"len([1])", 1},
		{"len([1, 1 + 2 * 3, true])", 3},
		// first for arrays
		{"first([])", nil},
		{"first([1])", 1},
		{"first([1, 2])", 1},
		{`first(1)`, "argument to `first` must be Array, got Integer"},
		// last for arrays
		{"last([])", nil},
		{"last([1])", 1},
		{"last([1, 2])", 2},
		{`last(1)`, "argument to `last` must be Array, got Integer"},
		// rest for arrays
		{"rest([])", nil},
		{"rest([1])", []int64{}},
		{"rest([1, 2, 3])", []int64{2, 3}},
		{`rest(1)`, "argument to `last` must be Array, got Integer"},
		// push for arrays
		{"push([], 1)", []int64{1}},
		{"push([1, 2], 3)", []int64{1, 2, 3}},
		{"push([])", "wrong number of arguments. want=2, got=1"},
		{"push(1, 2)", "first argument to `push` must be Array, got Integer"},
		// puts
		{"puts(1)", nil},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not *object.Error. got=%#v", evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
			}
		case []int64:
			arrObj, ok := evaluated.(*object.Array)
			if !ok {
				t.Errorf("object is not *object.Array. got=%#v", evaluated)
				continue
			}
			if len(arrObj.Elements) != len(expected) {
				t.Errorf("wrong number of elements. want=%d, got=%d",
					len(arrObj.Elements), len(expected))
				continue
			}
			for i, elem := range arrObj.Elements {
				testIntegerObject(t, elem, expected[i])
			}
		case nil:
			testNilObject(t, evaluated)
		default:
			t.Errorf("unsupported evaluated value: %#v, want=%#v", evaluated, tt.expected)
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(t, input)
	array, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not *object.Array. got=%#v", evaluated)
	}

	if l := len(array.Elements); l != 3 {
		t.Fatalf("array has wrong number of elements. want=%d, got=%d", 3, l)
	}

	testIntegerObject(t, array.Elements[0], 1)
	testIntegerObject(t, array.Elements[1], 4)
	testIntegerObject(t, array.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"[1, 2, 3][0]", 1},
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][2]", 3},
		{"let i = 0; [1][i]", 1},
		{"[1, 2, 3][1 + 1]", 3},
		{"let arr = [1, 2, 3]; arr[2];", 3},
		{"let arr = [1, 2, 3]; arr[0] + arr[1] + arr[2];", 6},
		{"let arr = [1, 2, 3]; let i = arr[0]; arr[i]", 2},
		{"[1, 2, 3][3]", nil},
		{"[1, 2, 3][-1]", nil},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		if i, ok := tt.expected.(int); ok {
			testIntegerObject(t, evaluated, int64(i))
			continue
		}
		testNilObject(t, evaluated)
	}
}

func TestHashLiterals(t *testing.T) {
	input := `
	let two = "two";
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4,
		true: 5,
		false: 6
	};
	`

	evaluated := testEval(t, input)
	hash, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("object is not *object.Hash. got=%#v", evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TrueValue.HashKey():                        5,
		FalseValue.HashKey():                       6,
	}

	if l := len(hash.Pairs); l != len(expected) {
		t.Fatalf("hash has wrong number of pairs. want=%d, got=%d", len(expected), l)
	}

	for key, value := range expected {
		pair, ok := hash.Pairs[key]
		if !ok {
			t.Errorf("no pair for given key in Pairs: %#v", key)
			continue
		}
		testIntegerObject(t, pair.Value, value)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`{"foo": 2 + 3}["foo"]`, 5},
		{`{"foo": 5}["bar"]`, nil},
		{`let key = "foo"; {"foo": 5}[key]`, 5},
		{`{}["foo"]`, nil},
		{`{5: 5}[5]`, 5},
		{`{true: 5}[true]`, 5},
		{`{false: 5}[false]`, 5},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		if i, ok := tt.expected.(int); ok {
			testIntegerObject(t, evaluated, int64(i))
			continue
		}
		testNilObject(t, evaluated)
	}
}

func TestQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`quote(5)`, `5`},
		{`quote(foobar)`, `foobar`},
		{`quote(foobar + barfoo)`, `(foobar + barfoo)`},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		q, ok := evaluated.(*object.Quote)
		if !ok {
			t.Fatalf("expected *object.Quote, but got %T (%#v)", evaluated, evaluated)
		}

		if q.Node == nil {
			t.Fatalf("quote.node is nil")
		}

		got := q.Node.String()
		if got != tt.want {
			t.Errorf("expected %q, but got %q", tt.want, got)
		}
	}
}
