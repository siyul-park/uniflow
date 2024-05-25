package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/siyul-park/uniflow/pkg/script/token"
)

// Node represents an AST node.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement.
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression.
type Expression interface {
	Node
	expressionNode()
}

// Program is a top-level AST node of a program.
type Program struct {
	Statements []Statement
}

// TokenLiteral returns the first token literal of a program.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) == 0 {
		return ""
	}
	return p.Statements[0].TokenLiteral()
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// LetStatement represents a let statement.
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Ident
	Value Expression
}

func (ls *LetStatement) statementNode() {}

// TokenLiteral returns a token literal of let statement.
func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

// AssignStatement represents an assignment statement.
type AssignStatement struct {
	Token    token.Token // token.ASSIGN
	LHS, RHS Expression
}

func (as *AssignStatement) statementNode() {}

// TokenLiteral returns a token literal of assignment statement.
func (as *AssignStatement) TokenLiteral() string {
	return as.Token.Literal
}

func (as *AssignStatement) String() string {
	var out strings.Builder

	// Left-hand side expression
	out.WriteString(as.LHS.String())

	// Assignment symbol (equal sign)
	out.WriteString(" = ")

	// Right-hand side expression
	if as.RHS != nil {
		out.WriteString(as.RHS.String())
	}

	out.WriteString(";")

	return out.String()
}

// Ident represents an identifier.
type Ident struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Ident) expressionNode() {}

// TokenLiteral returns a token literal of an identifier.
func (i *Ident) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Ident) String() string {
	return i.Value
}

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Token       token.Token // the token.RETURN token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}

// TokenLiteral returns a token literal of return statement.
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

// ExpressionStatement represents an expression statement.
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

// TokenLiteral returns a token literal of expression statement.
func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}

func (es *ExpressionStatement) String() string {
	if es.Expression == nil {
		return ""
	}
	return es.Expression.String()
}

// IntegerLiteral represents an integer literal.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}

// TokenLiteral returns a token literal of integer.
func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
}

func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

// FloatLiteral represents a floating point number literal.
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode() {}

// TokenLiteral returns a token literal of floating point number.
func (fl *FloatLiteral) TokenLiteral() string {
	return fl.Token.Literal
}

func (fl *FloatLiteral) String() string {
	return fl.Token.Literal
}

// PrefixExpression represents a prefix expression.
type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}

// TokenLiteral returns a token literal.
func (pe *PrefixExpression) TokenLiteral() string {
	return pe.Token.Literal
}

func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

// InfixExpression represents an infix expression.
type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}

// TokenLiteral returns a token literal.
func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}

func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

// Boolean represents a boolean value.
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode() {}

// TokenLiteral returns a token literal of boolean value.
func (b *Boolean) TokenLiteral() string {
	return b.Token.Literal
}

func (b *Boolean) String() string {
	return b.TokenLiteral()
}

// Nil represents nil value.
type Nil struct {
	Token token.Token
}

func (n *Nil) expressionNode() {}

// TokenLiteral returns a token literal of nil value.
func (n *Nil) TokenLiteral() string {
	return n.Token.Literal
}

func (n *Nil) String() string {
	return n.TokenLiteral()
}

// IfExpression represents an if expression.
type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode() {}

// TokenLiteral returns a token literal of if expression.
func (ie *IfExpression) TokenLiteral() string {
	return ie.Token.Literal
}

func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

// BlockStatement represents a block statement.
type BlockStatement struct {
	Token      token.Token // the '{' token
	Statements []Statement
}

func (bs *BlockStatement) expressionNode() {}

// TokenLiteral returns a token literal of block statement.
func (bs *BlockStatement) TokenLiteral() string {
	return bs.Token.Literal
}

func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// FunctionLiteral represents a fuction literal.
type FunctionLiteral struct {
	Token      token.Token
	Parameters []*Ident
	Body       *BlockStatement
	Name       string
}

func (fl *FunctionLiteral) expressionNode() {}

// TokenLiteral returns a token literal of function.
func (fl *FunctionLiteral) TokenLiteral() string {
	return fl.Token.Literal
}

func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := make([]string, 0, len(fl.Parameters))
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	if fl.Name != "" {
		out.WriteString(fmt.Sprintf("<%s>", fl.Name))
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

// CallExpression represents a function call expression.
type CallExpression struct {
	Token     token.Token // the '(' token
	Function  Expression  // Ident or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}

// TokenLiteral returns a token literal of function call expression.
func (ce *CallExpression) TokenLiteral() string {
	return ce.Token.Literal
}

func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := make([]string, 0, len(ce.Arguments))
	for _, arg := range ce.Arguments {
		args = append(args, arg.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

// StringLiteral represents a string literal.
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}

// TokenLiteral returns a token literal of string.
func (sl *StringLiteral) TokenLiteral() string {
	if sl == nil {
		return ""
	}
	return sl.Token.Literal
}

func (sl *StringLiteral) String() string {
	return sl.TokenLiteral()
}

// ArrayLiteral represents an array literal.
type ArrayLiteral struct {
	Token    token.Token // the '[' token
	Elements []Expression
}

func (*ArrayLiteral) expressionNode() {}

// TokenLiteral returns a token literal of array.
func (al *ArrayLiteral) TokenLiteral() string {
	if al == nil {
		return ""
	}
	return al.Token.Literal
}

func (al *ArrayLiteral) String() string {
	if al == nil {
		return ""
	}

	elements := make([]string, 0, len(al.Elements))
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	var out bytes.Buffer

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// IndexExpression represents an expression in array index operator.
type IndexExpression struct {
	Token token.Token // the '[' token
	Left  Expression
	Index Expression
}

func (*IndexExpression) expressionNode() {}

// TokenLiteral returns a token literal of array.
func (ie *IndexExpression) TokenLiteral() string {
	if ie == nil {
		return ""
	}
	return ie.Token.Literal
}

func (ie *IndexExpression) String() string {
	if ie == nil {
		return ""
	}

	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

// HashLiteral represents a hash literal.
type HashLiteral struct {
	Token token.Token // the '{' token
	Pairs map[Expression]Expression
}

func (*HashLiteral) expressionNode() {}

// TokenLiteral returns a token literal of hash.
func (hl *HashLiteral) TokenLiteral() string {
	if hl == nil {
		return ""
	}
	return hl.Token.Literal
}

func (hl *HashLiteral) String() string {
	if hl == nil {
		return ""
	}

	pairs := make([]string, len(hl.Pairs))
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+": "+value.String())
	}

	var out bytes.Buffer
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// MacroLiteral represents a macro literal.
type MacroLiteral struct {
	Token      token.Token
	Parameters []*Ident
	Body       *BlockStatement
}

func (ml *MacroLiteral) expressionNode() {}

// TokenLiteral returns a token literal of function.
func (ml *MacroLiteral) TokenLiteral() string {
	return ml.Token.Literal
}

func (ml *MacroLiteral) String() string {
	var out bytes.Buffer

	params := make([]string, 0, len(ml.Parameters))
	for _, p := range ml.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(ml.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(ml.Body.String())

	return out.String()
}
