package eval

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/script/ast"
	"github.com/siyul-park/uniflow/pkg/script/lexer"
	"github.com/siyul-park/uniflow/pkg/script/object"
	"github.com/siyul-park/uniflow/pkg/script/parser"
)

func TestDefineMacros(t *testing.T) {
	input := `
  let num = 1;
	let func = fn(x, y) { x + y };
	let mymacro = macro(x, y) { x + y; };
	`

	env := object.NewEnvironment()
	program := testParseProgram(input)

	DefineMacros(program, env)

	stmts := program.Statements
	if len(stmts) != 2 {
		t.Fatalf("Wrong number of statements. got=%d", len(stmts))
	}

	if _, ok := env.Get("num"); ok {
		t.Fatalf("`num` variable should not be defined")
	}

	if _, ok := env.Get("func"); ok {
		t.Fatalf("`func` variable should not be defined")
	}

	obj, ok := env.Get("mymacro")
	if !ok {
		t.Fatalf("mymacro not in the environment")
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		t.Fatalf("object is not Macro; got %T (%#v)", obj, obj)
	}

	params := macro.Parameters
	if len(params) != 2 {
		t.Fatalf("Wrong number of macro parameters; got %d", len(params))
	}

	if params[0].String() != "x" {
		t.Fatalf("parameter is not 'x'; got %q", params[0])
	}
	if params[1].String() != "y" {
		t.Fatalf("parameter is not 'y'; got %q", params[1])
	}

	want := "(x + y)"

	got := macro.Body.String()
	if got != want {
		t.Errorf("expected body %q but got %q", want, got)
	}
}

func TestExpandMacros(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: `
			let infixExpr = macro() { quote(1 + 2); };
			infixExpr();
			`,
			want: `(1 + 2)`,
		},
		{
			input: `
			let reverse = macro(a, b) { quote(unquote(b) - unquote(a)); };
			reverse(2 + 2, 10 - 5);
			`,
			want: `(10 - 5) - (2 + 2)`,
		},
		{
			input: `
			let unless = macro(condition, consequence, altenative) {
			  quote(
					if (!(unquote(condition))) {
						unquote(consequence)
					} else {
						unquote(altenative)
					}
				);
			};

			unless(10 > 5, puts("not greater"), puts("greater"));
			`,
			want: `
			if (!(10 > 5)) {
				puts("not greater")
			} else {
				puts("greater")
			}
			`,
		},
	}

	for _, tt := range tests {
		program := testParseProgram(tt.input)
		env := object.NewEnvironment()
		DefineMacros(program, env)
		got := ExpandMacros(program, env).String()

		want := testParseProgram(tt.want).String()
		if got != want {
			t.Errorf("expected %q, but got %q", want, got)
		}
	}
}

func testParseProgram(input string) *ast.Program {
	return parser.New(lexer.New(input)).ParseProgram()
}
