package parser

import (
	"fmt"
	"testing"

	"github.com/siyul-park/uniflow/pkg/script/ast"
	"github.com/siyul-park/uniflow/pkg/script/lexer"
)

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedIdent string
		expectedValue interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))

		program := p.ParseProgram()
		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}
		l := len(program.Statements)
		if l != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, l)
		}
		checkParserErrors(t, p)

		stmt := program.Statements[0]
		testLetStatement(t, stmt, tt.expectedIdent)

		val := stmt.(*ast.LetStatement).Value
		testLiteralExpression(t, val, tt.expectedValue)
	}
}

func TestLetStatementErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"let = 5;"},
		{"let x = ;"},
		{"let x 1;"},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))

		program := p.ParseProgram()
		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}

		if len(p.Errors()) == 0 {
			t.Errorf("parser has no errors despite invalid statements.")
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("s.Name not '%s'. got=%s", name, letStmt.Name)
	}
}

func TestAssignmentStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedIdent string
		expectedValue interface{}
	}{
		{"x = 5;", "x", 5},
		{"y = true;", "y", true},
		{"foobar = y;", "foobar", "y"},
		{"x = 5; x = 6;", "x", 6},
		{"a = 5.5; b = 6.6;", "b", 6.6},
		{"y = true; y = false;", "y", false},
		{"foobar = y; foobar = z;", "foobar", "z"},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))

		program := p.ParseProgram()
		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}
		l := len(program.Statements)
		if l < 1 && l > 2 {
			t.Fatalf("program.Statements does not contain one or two statements. got=%d", l)
		}
		checkParserErrors(t, p)

		// Check the last statement
		stmt := program.Statements[l-1]
		testAssignmentStatement(t, stmt, tt.expectedIdent)

		val := stmt.(*ast.AssignStatement).RHS
		testLiteralExpression(t, val, tt.expectedValue)
	}
}

func testAssignmentStatement(t *testing.T, s ast.Statement, name string) {
	stmt, ok := s.(*ast.AssignStatement)
	if !ok {
		t.Errorf("statement not *ast.AssignmentStatement. got=%T", s)
	}

	ident, ok := stmt.LHS.(*ast.Ident)
	if !ok {
		t.Errorf("stmt.LHS not identifier. got=%T (%#v)", ident, ident)
	}

	if got := ident.Value; got != name {
		t.Errorf("stmt.Name.Value not %q. got=%q", name, got)
	}

	if got := ident.TokenLiteral(); got != name {
		t.Errorf("stmt.Name not %q. got=%q", name, got)
	}
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return true;", true},
		{"return x;", "x"},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))

		program := p.ParseProgram()
		l := len(program.Statements)
		if l != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, l)
		}
		checkParserErrors(t, p)

		for _, stmt := range program.Statements {
			returnStmt, ok := stmt.(*ast.ReturnStatement)
			if !ok {
				t.Errorf("stmt not *ast.returnStmt. got=%T", stmt)
				continue
			}
			if returnStmt.ReturnValue.String() != fmt.Sprintf("%v", tt.expectedValue) {
				t.Errorf("returnStmt.ReturnValue not %v, got %v", tt.expectedValue, returnStmt.ReturnValue)
			}
			if returnStmt.TokenLiteral() != "return" {
				t.Errorf("returnStmt.TokenLiteral not 'return', got %q", returnStmt.TokenLiteral())
			}
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar"

	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	l := len(program.Statements)
	if l != 1 {
		t.Fatalf("program has not enough statements. got=%d", l)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	testIdent(t, stmt.Expression, input)
}

func TestIntegerExpression(t *testing.T) {
	input := "5;"

	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	l := len(program.Statements)
	if l != 1 {
		t.Fatalf("program has not enough statements. got=%d", l)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	testIntegerLiteral(t, stmt.Expression, 5)
}

func testFloatLiteral(t *testing.T, expr ast.Expression, value float64) {
	fl, ok := expr.(*ast.FloatLiteral)
	if !ok {
		t.Errorf("expr not *ast.FloatLiteral. got=%T", fl)
		return
	}

	if fl.Value != value {
		t.Errorf("fl.Value not %f. got=%f", value, fl.Value)
	}
}

func TestFloatExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"12.34", 12.34},
		{"0.56", 0.56},
		{"78.00", 78.00},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if l := len(program.Statements); l != 1 {
			t.Errorf("program has not enough statements. got=%d", l)
			continue
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
				program.Statements[0])
			continue
		}

		testFloatLiteral(t, stmt.Expression, tt.expected)
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		want     interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"-15.5;", "-", 15.5},
		{"!true", "!", true},
		{"!false", "!", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		length := len(program.Statements)
		if length != 1 {
			t.Fatalf("program has not enough statements. got=%d", length)
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("exp not *ast.PrefixExpression. got=%T", stmt.Expression)
		}
		if exp.Operator != tt.operator {
			t.Errorf("exp.operator is not %s. got=%s", tt.operator, exp.Operator)
		}

		testLiteralExpression(t, exp.Right, tt.want)
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) {
	i, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("il not *ast.IntegerLiteral. got=%T", il)
	}

	if i.Value != value {
		t.Errorf("i.Value not %d. got=%d", value, i.Value)
	}

	if i.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("i.TokenLiteral() not %d. got=%s", value, i.TokenLiteral())
	}
}

func testIdent(t *testing.T, expr ast.Expression, value string) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		t.Errorf("expr not *ast.Ident. got=%T", expr)
	}
	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral() not %s. got=%s", value, ident.TokenLiteral())
	}
}

func testLiteralExpression(t *testing.T, expr ast.Expression, expected interface{}) {
	switch v := expected.(type) {
	case int:
		testIntegerLiteral(t, expr, int64(v))
	case int64:
		testIntegerLiteral(t, expr, v)
	case float64:
		testFloatLiteral(t, expr, v)
	case string:
		testIdent(t, expr, v)
	case bool:
		testBooleanLiteral(t, expr, v)
	default:
		t.Errorf("type of expr not handled. got=%T", expr)
	}
}

func testInfixExpression(t *testing.T, expr ast.Expression, left interface{}, operator string,
	right interface{}) {
	op, ok := expr.(*ast.InfixExpression)
	if !ok {
		t.Errorf("expr is not ast.OperatorExpression. got=%T(%s)", expr, expr)
		return
	}

	testLiteralExpression(t, op.Left, left)

	if op.Operator != operator {
		t.Errorf("expr.Operator is not %q. got=%q", operator, op.Operator)
	}

	testLiteralExpression(t, op.Right, right)
}

func TestParsingInfixExpressions(t *testing.T) {
	tests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 >= 5;", 5, ">=", 5},
		{"5 <= 5;", 5, "<=", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
		{"true && true", true, "&&", true},
		{"true || false", true, "||", false},
		{"0 && 1", 0, "&&", 1},
		{"0 || 1", 0, "||", 1},
		{"5 + 5.1;", 5, "+", 5.1},
		{"5.0 - 5.2;", 5.0, "-", 5.2},
		{"5.3 * 5.4;", 5.3, "*", 5.4},
		{"5.5 / 5.6;", 5.5, "/", 5.6},
		{"5.7 > 5.8;", 5.7, ">", 5.8},
		{"5.9 < 5;", 5.9, "<", 5},
		{"5.7 >= 5.8;", 5.7, ">=", 5.8},
		{"5.9 <= 5;", 5.9, "<=", 5},
		{"5 == 5.0;", 5, "==", 5.0},
		{"5.1 != 5.1;", 5.1, "!=", 5.1},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))
		program := p.ParseProgram()
		checkParserErrors(t, p)

		l := len(program.Statements)
		if l != 1 {
			t.Errorf("program.Statements does not contain %d statements. got=%d", 1, l)
			continue
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
				program.Statements[0])
			continue
		}

		expr, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Errorf("expr not *ast.InfixExpression. got=%T", stmt.Expression)
			continue
		}

		testLiteralExpression(t, expr.Left, tt.leftValue)

		if expr.Operator != tt.operator {
			t.Errorf("expr.Operator is not %q. got=%s", tt.operator, expr.Operator)
		}

		testLiteralExpression(t, expr.Right, tt.rightValue)
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b - c", "((a + b) - c)"},
		{"a * b * c", "((a * b) * c)"},
		{"a * b / c", "((a * b) / c)"},
		{"a + b / c", "(a + (b / c))"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"3 + 4; -5 * 5", "(3 + 4)((-5) * 5)"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4", "((5 < 4) != (3 > 4))"},
		{"5 >= 4 == 3 <= 4", "((5 >= 4) == (3 <= 4))"},
		{"5 <= 4 != 3 >= 4", "((5 <= 4) != (3 >= 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"true", "true"},
		{"false", "false"},
		{"3 > 5 == false", "((3 > 5) == false)"},
		{"3 < 5 == true", "((3 < 5) == true)"},
		{"3 >= 5 == false", "((3 >= 5) == false)"},
		{"3 <= 5 == true", "((3 <= 5) == true)"},
		{"3 > 5 && false", "((3 > 5) && false)"},
		{"3 < 5 && true", "((3 < 5) && true)"},
		{"3 + 5 || false", "((3 + 5) || false)"},
		{"3 * 5 || true", "((3 * 5) || true)"},
		{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4)"},
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
		{"2 / (5 + 5)", "(2 / (5 + 5))"},
		{"-(5 + 5)", "(-(5 + 5))"},
		{"!(true == true)", "(!(true == true))"},
		{"!(true && true)", "(!(true && true))"},
		{"a + add(b * c) + d", "((a + add((b * c))) + d)"},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{"add(a + b + c * d / f + g)", "add((((a + b) + ((c * d) / f)) + g))"},
		{"a * [1, 2, 3, 4][b * c] * d", "((a * ([1, 2, 3, 4][(b * c)])) * d)"},
		{"add(a * b[2], b[1], 2 * [1, 2][1])", "add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))"},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))

		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func testBooleanLiteral(t *testing.T, expr ast.Expression, value bool) {
	b, ok := expr.(*ast.Boolean)
	if !ok {
		t.Errorf("b not *ast.Boolean. got=%T", expr)
	}
	if b.Value != value {
		t.Errorf("b.Value not %t. got=%t", value, b.Value)
	}
	if b.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("b.TokenLiteral() not %t. got=%s", value, b.TokenLiteral())
	}
}

func TestNilExpressions(t *testing.T) {
	input := "nil"

	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	l := len(program.Statements)
	if l != 1 {
		t.Fatalf("program has not enough statements. got=%d", l)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	if _, ok := stmt.Expression.(*ast.Nil); !ok {
		t.Errorf("stmt is not *ast.Nil")
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))
		program := p.ParseProgram()
		checkParserErrors(t, p)

		l := len(program.Statements)
		if l != 1 {
			t.Fatalf("program has not enough statements. got=%d", l)
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		testBooleanLiteral(t, stmt.Expression, tt.expected)
	}
}

func TestIfExpression(t *testing.T) {
	input := "if (x < y) { x }"

	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	l := len(program.Statements)
	if l != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d", 1, l)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	expr, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Errorf("stmt.Expression is not ast.Expression. got=%T", stmt.Expression)
	}

	testInfixExpression(t, expr.Condition, "x", "<", "y")

	l = len(expr.Consequence.Statements)
	if l != 1 {
		t.Errorf("consequence is not %d statements. got=%d\n", 1, l)
	}

	cons, ok := expr.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not *ast.ExpressionStatement. got=%T",
			expr.Consequence.Statements[0])
	}

	testIdent(t, cons.Expression, "x")

	if expr.Alternative != nil {
		t.Errorf("expr.Alternative.Statements was not nil. got=%+v", expr.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := "if (x < y) { x } else { y }"

	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	l := len(program.Statements)
	if l != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d", 1, l)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	expr, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Errorf("stmt.Expression is not ast.Expression. got=%T", stmt.Expression)
	}

	testInfixExpression(t, expr.Condition, "x", "<", "y")

	l = len(expr.Consequence.Statements)
	if l != 1 {
		t.Errorf("consequence is not %d statements. got=%d\n", 1, l)
	}

	cons, ok := expr.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not *ast.ExpressionStatement. got=%T",
			expr.Consequence.Statements[0])
	}

	testIdent(t, cons.Expression, "x")

	alt, ok := expr.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not *ast.ExpressionStatement. got=%T",
			expr.Alternative.Statements[0])
	}

	testIdent(t, alt.Expression, "y")
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := "fn(x, y) { x + y; }"

	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	l := len(program.Statements)
	if l != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d", 1, l)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	f, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Errorf("stmt.Expression is not *ast.FunctionLiteral. got=%T", stmt.Expression)
	}

	l = len(f.Parameters)
	if l != 2 {
		t.Fatalf("function literal parameters wrong. want=%d, got=%d\n", 2, l)
	}

	testLiteralExpression(t, f.Parameters[0], "x")
	testLiteralExpression(t, f.Parameters[1], "y")

	l = len(f.Body.Statements)
	if l != 1 {
		t.Fatalf("f.Body.Statements has not %d statements. got=%d\n", 1, l)
	}

	bodyStmt, ok := f.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("f.Body.Statements[0] is not *ast.ExpressionStatement. got=%T", f.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionLiteralWithName(t *testing.T) {
	input := "let myFunc = fn() { };"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmtLen := len(program.Statements)
	if stmtLen != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d", 1, stmtLen)
	}

	stmt, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.LetStatement. got=%T", program.Statements[0])
	}

	fn, ok := stmt.Value.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Value is not *ast.FunctionLiteral. got=%T", stmt.Value)
	}

	if fn.Name != "myFunc" {
		t.Fatalf("function literal name wrong. want=%q, got=%q", "myFunc", fn.Name)
	}
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"fn() {}", []string{}},
		{"fn(x) {};", []string{"x"}},
		{"fn(x, y, z) {};", []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		f := stmt.Expression.(*ast.FunctionLiteral)

		if len(f.Parameters) != len(tt.expected) {
			t.Errorf("length parameters wrong. want=%d, got=%d", len(tt.expected), len(f.Parameters))
		}

		for i, ident := range tt.expected {
			testLiteralExpression(t, f.Parameters[i], ident)
		}
	}
}

func TestCallFunctionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"

	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	l := len(program.Statements)
	if l != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, l)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	expr, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Errorf("stmt.Expression is not *ast.CallExpression. got=%T", stmt.Expression)
	}

	testIdent(t, expr.Function, "add")

	l = len(expr.Arguments)
	if l != 3 {
		t.Fatalf("wrong length of arguments. got=%d\n", l)
	}

	testLiteralExpression(t, expr.Arguments[0], 1)
	testInfixExpression(t, expr.Arguments[1], 2, "*", 3)
	testInfixExpression(t, expr.Arguments[2], 4, "+", 5)
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if l := len(program.Statements); l != 1 {
		t.Fatalf("program has not 1 statement. got=%d", l)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("literal not *ast.StringLiteral. got=%T", stmt.Expression)
	}

	expected := "hello world"
	if literal.Value != expected {
		t.Errorf("literal.Value not %q. got=%q", expected, literal.Value)
	}
}

func TestParsingArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if l := len(program.Statements); l != 1 {
		t.Fatalf("program has not 1 statement. got=%d", l)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("array not *ast.ArrayLiteral. got=%T", stmt.Expression)
	}

	if l := len(array.Elements); l != 3 {
		t.Fatalf("len(array.Elements) not %d. got=%d", 3, l)
	}

	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestParsingIndexExpressions(t *testing.T) {
	input := "myArray[1 + 1]"

	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if l := len(program.Statements); l != 1 {
		t.Fatalf("program has not %d statement. got=%d", 1, l)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	idxExpr, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("idxExpr not *ast.IndexExpression. got=%T", stmt.Expression)
	}

	testIdent(t, idxExpr.Left, "myArray")
	testInfixExpression(t, idxExpr.Index, 1, "+", 1)
}

func TestParsingHashLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			input:    "{}",
			expected: map[string]int64{},
		},
		{
			input: `{"one": 1, "two": 2, "three": 3}`,
			expected: map[string]int64{
				"one":   1,
				"two":   2,
				"three": 3,
			},
		},
		{
			input: "{1: 1, 2: 2, 3: 3}",
			expected: map[int64]int64{
				1: 1,
				2: 2,
				3: 3,
			},
		},
		{
			input: "{true: 1, false: 2}",
			expected: map[bool]int64{
				true:  1,
				false: 2,
			},
		},
		{
			input: `{"one": 0 + 1, "two": 10 - 8, "three": 15 / 5}`,
			expected: map[string]func(ast.Expression){
				"one": func(e ast.Expression) {
					testInfixExpression(t, e, 0, "+", 1)
				},
				"two": func(e ast.Expression) {
					testInfixExpression(t, e, 10, "-", 8)
				},
				"three": func(e ast.Expression) {
					testInfixExpression(t, e, 15, "/", 5)
				},
			},
		},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if l := len(program.Statements); l != 1 {
			t.Fatalf("program has not 1 statement. got=%d", l)
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		hash, ok := stmt.Expression.(*ast.HashLiteral)
		if !ok {
			t.Fatalf("hash not *ast.HashLiteral. got=%T", stmt.Expression)
		}

		for key, value := range hash.Pairs {
			switch key := key.(type) {
			case *ast.StringLiteral:
				switch expected := tt.expected.(type) {
				case map[string]int64:
					if l := len(hash.Pairs); l != len(expected) {
						t.Errorf("hash.Pairs has wrong length. want=%d, got=%d", len(expected), l)
						continue
					}
					expectedValue := expected[key.Value]
					testIntegerLiteral(t, value, expectedValue)
				case map[string]func(ast.Expression):
					if l := len(hash.Pairs); l != len(expected) {
						t.Errorf("hash.Pairs has wrong length. want=%d, got=%d", len(expected), l)
						continue
					}
					testFunc := expected[key.Value]
					testFunc(value)
				}
			case *ast.IntegerLiteral:
				expected := tt.expected.(map[int64]int64)
				if l := len(hash.Pairs); l != len(expected) {
					t.Errorf("hash.Pairs has wrong length. want=%d, got=%d", len(expected), l)
					continue
				}
				expectedValue := expected[key.Value]
				testIntegerLiteral(t, value, expectedValue)
			case *ast.Boolean:
				expected := tt.expected.(map[bool]int64)
				if l := len(hash.Pairs); l != len(expected) {
					t.Errorf("hash.Pairs has wrong length. want=%d, got=%d", len(expected), l)
					continue
				}
				expectedValue := expected[key.Value]
				testIntegerLiteral(t, value, expectedValue)
			default:
				t.Errorf("unsupported key type: %T", key)
			}
		}
	}
}

func TestMacroLiteralParsing(t *testing.T) {
	input := `macro(x, y) { x + y; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmts := program.Statements
	if len(stmts) != 1 {
		t.Fatalf("expect program.Statements to contain 1 statements, but got %d statements", len(stmts))
	}

	stmt, ok := stmts[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not *ast.ExpressionStatement; got %T", stmts[0])
	}

	macro, ok := stmt.Expression.(*ast.MacroLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.MacroLiteral; got %T", stmt.Expression)
	}

	params := macro.Parameters
	if len(params) != 2 {
		t.Fatalf("macro literal parameters wrong. want 2, but got %d", len(params))
	}

	testLiteralExpression(t, params[0], "x")
	testLiteralExpression(t, params[1], "y")

	stmts = macro.Body.Statements
	if len(stmts) != 1 {
		t.Fatalf("macro.Body.Statements has not 1 statements; got %d statements", len(stmts))
	}

	bodyStmt, ok := stmts[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("macro body stmt is not *ast.ExpressionStatement; got %T", stmts[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	length := len(errors)
	if length == 0 {
		return
	}

	t.Errorf("parser has %d errors", length)
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
