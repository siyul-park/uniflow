package ast

import (
	"reflect"
	"testing"
)

func createIntLitFunc(i int64) func() Expression {
	return func() Expression { return &IntegerLiteral{Value: i} }
}

func TestModify(t *testing.T) {
	one := createIntLitFunc(1)
	two := createIntLitFunc(2)

	turnOneIntoTwo := func(node Node) Node {
		i, ok := node.(*IntegerLiteral)
		if !ok {
			return node
		}

		if i.Value != 1 {
			return node
		}

		i.Value = 2
		return i
	}

	tests := []struct {
		input Node
		want  Node
	}{
		{
			input: one(),
			want:  two(),
		},
		{
			input: &Program{Statements: []Statement{&ExpressionStatement{Expression: one()}}},
			want:  &Program{Statements: []Statement{&ExpressionStatement{Expression: two()}}},
		},
		{
			input: &InfixExpression{Left: one(), Operator: "+", Right: two()},
			want:  &InfixExpression{Left: two(), Operator: "+", Right: two()},
		},
		{
			input: &InfixExpression{Left: two(), Operator: "+", Right: one()},
			want:  &InfixExpression{Left: two(), Operator: "+", Right: two()},
		},
		{
			input: &PrefixExpression{Operator: "-", Right: one()},
			want:  &PrefixExpression{Operator: "-", Right: two()},
		},
		{
			input: &IndexExpression{Left: one(), Index: one()},
			want:  &IndexExpression{Left: two(), Index: two()},
		},
		{
			input: &IfExpression{
				Condition: one(),
				Consequence: &BlockStatement{
					Statements: []Statement{&ExpressionStatement{Expression: one()}},
				},
				Alternative: &BlockStatement{
					Statements: []Statement{&ExpressionStatement{Expression: one()}},
				},
			},
			want: &IfExpression{
				Condition: two(),
				Consequence: &BlockStatement{
					Statements: []Statement{&ExpressionStatement{Expression: two()}},
				},
				Alternative: &BlockStatement{
					Statements: []Statement{&ExpressionStatement{Expression: two()}},
				},
			},
		},
		{
			input: &ReturnStatement{ReturnValue: one()},
			want:  &ReturnStatement{ReturnValue: two()},
		},
		{
			input: &LetStatement{Value: one()},
			want:  &LetStatement{Value: two()},
		},
		{
			input: &FunctionLiteral{
				Parameters: []*Ident{},
				Body: &BlockStatement{
					Statements: []Statement{&ExpressionStatement{Expression: one()}},
				},
			},
			want: &FunctionLiteral{
				Parameters: []*Ident{},
				Body: &BlockStatement{
					Statements: []Statement{&ExpressionStatement{Expression: two()}},
				},
			},
		},
		{
			input: &ArrayLiteral{Elements: []Expression{one(), one()}},
			want:  &ArrayLiteral{Elements: []Expression{two(), two()}},
		},
	}

	for _, tt := range tests {
		got := Modify(tt.input, turnOneIntoTwo)

		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("expected %#v, but got %#v", tt.want, got)
		}
	}

	// Test for hash literals

	hashLit := &HashLiteral{
		Pairs: map[Expression]Expression{
			one(): one(),
			one(): one(),
		},
	}

	Modify(hashLit, turnOneIntoTwo)

	for key, val := range hashLit.Pairs {
		key := key.(*IntegerLiteral)
		if key.Value != 2 {
			t.Errorf("key is not %d and got %d", 2, key.Value)
		}
		val := val.(*IntegerLiteral)
		if val.Value != 2 {
			t.Errorf("value is not %d and got %d", 2, key.Value)
		}
	}
}
