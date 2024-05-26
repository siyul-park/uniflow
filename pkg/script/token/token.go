package token

// Type is a token type.
type Type string

const (
	// ILLEGAL is a token type for illegal tokens.
	ILLEGAL Type = "ILLEGAL"
	// EOF is a token type that represents end of file.
	EOF = "EOF"

	// IDENT is a token type for identifiers.
	IDENT = "IDENT" // add, foobar, x, y, ...
	// INT is a token type for integers.
	INT = "INT"
	// FLOAT is a token type for floating point numbers.
	FLOAT = "FLOAT"
	// STRING is a token type for strings.
	STRING = "STRING"

	// BANG is a token type for NOT operator.
	BANG = "!"
	// ASSIGN is a token type for assignment operators.
	ASSIGN = "="
	// PLUS is a token type for addition.
	PLUS = "+"
	// MINUS is a token type for subtraction.
	MINUS = "-"
	// ASTARISK is a token type for multiplication.
	ASTARISK = "*"
	// SLASH is a token type for division.
	SLASH = "/"
	// LT is a token ype for 'less than' operator.
	LT = "<"
	// GT is a token ype for 'greater than' operator.
	GT = ">"
	// LE is a token type for 'less than or equal to' operator.
	LE = "<="
	// GE is a token type for 'greater than or equal to' operator.
	GE = ">="
	// EQ is a token type for equality operator.
	EQ = "=="
	// NEQ is a token type for not equality operator.
	NEQ = "!="
	// AND is a token type for binary AND logical operator.
	AND = "&&"
	// OR is a token type for binary OR logical operator.
	OR = "||"

	// COMMA is a token type for commas.
	COMMA = ","
	// SEMICOLON is a token type for semicolons.
	SEMICOLON = ";"
	// COLON is a token type for colons.
	COLON = ":"

	// LPAREN is a token type for left parentheses.
	LPAREN = "("
	// RPAREN is a token type for right parentheses.
	RPAREN = ")"
	// LBRACE is a token type for left braces.
	LBRACE = "{"
	// RBRACE is a token type for right braces.
	RBRACE = "}"
	// LBRACKET is a token type for left brackets.
	LBRACKET = "["
	// RBRACKET is a token type for right brackets.
	RBRACKET = "]"

	// FUNCTION is a token type for functions.
	FUNCTION = "FUNCTION"
	// LET is a token type for lets.
	LET = "LET"
	// TRUE is a token type for true.
	TRUE = "TRUE"
	// FALSE is a token type for false.
	FALSE = "FALSE"
	// NIL is a token type for nil.
	NIL = "NIL"
	// IF is a token type for if.
	IF = "IF"
	// ELSE is a token type for else.
	ELSE = "ELSE"
	// RETURN is a token type for return.
	RETURN = "RETURN"
	// MACRO is a token type for macros.
	MACRO = "MACRO"
)

// Token represents a token which has a token type and literal.
type Token struct {
	Type    Type
	Literal string
}

// Language keywords
var keywords = map[string]Type{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"nil":    NIL,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"macro":  MACRO,
}

// LookupIdent checks the language keywords to see whether the given identifier is a keyword.
// If it is, it returns the keyword's Type constant. If it isn't, it just gets back IDENT.
func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
