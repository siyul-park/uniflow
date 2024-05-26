package eval

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/script/object"
)

func TestQuoteUnquote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			`quote(unquote(4))`,
			`4`,
		},
		{
			`quote(unquote(4 + 4))`,
			`8`,
		},
		{
			`quote(8 + unquote(4 + 4))`,
			`(8 + 8)`,
		},
		{
			`quote(unquote(4 + 4) + 8)`,
			`(8 + 8)`,
		},
		{
			`let foobar = 8; quote(foobar)`,
			`foobar`,
		},
		{
			`let foobar = 8; quote(unquote(foobar))`,
			`8`,
		},
		{
			`quote(unquote(true))`,
			`true`,
		},
		{
			`quote(unquote(true == false))`,
			`false`,
		},
		{
			`quote(unquote(quote(4 + 4)))`,
			`(4 + 4)`,
		},
		{
			`let quotedInfixExpr = quote(4 + 4);
		   quote(unquote(4 + 4) + unquote(quotedInfixExpr))`,
			`(8 + (4 + 4))`,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		q, ok := evaluated.(*object.Quote)
		if !ok {
			t.Fatalf("expected *object.Quote, but got %T (%#v)", evaluated, evaluated)
		}

		if q.Node == nil {
			t.Fatalf("quote.Node is nil")
		}

		got := q.Node.String()
		if got != tt.want {
			t.Errorf("expected %q, but got %q", tt.want, got)
		}
	}
}
