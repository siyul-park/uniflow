package lexer

import "github.com/siyul-park/uniflow/pkg/script/token"

// Lexer represents a lexer for Monkey programming language.
type Lexer interface {
	// NextToken returns a next token.
	NextToken() token.Token
}

type lexer struct {
	input string
	// current position in input (points to current char)
	position int
	// current reading position in input (after current char)
	readPosition int
	// current char under examination
	ch byte
}

// New returns a new Lexer.
func New(input string) Lexer {
	l := &lexer{input: input}
	l.readChar()
	return l
}

func (l *lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *lexer) NextToken() token.Token {
	l.skipWhitespace()

	// skip comments
	if l.ch == '#' {
		l.skipComment()
	}

	var tok token.Token
	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			tok = l.readTwoCharToken(token.EQ)
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			tok = l.readTwoCharToken(token.NEQ)
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '*':
		tok = newToken(token.ASTARISK, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '<':
		if l.peekChar() == '=' {
			tok = l.readTwoCharToken(token.LE)
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			tok = l.readTwoCharToken(token.GE)
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			tok = l.readTwoCharToken(token.AND)
		}
	case '|':
		if l.peekChar() == '|' {
			tok = l.readTwoCharToken(token.OR)
		}
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isDigit(l.ch) {
			return l.readNumberToken()
		}

		if isLetter(l.ch) {
			tok.Literal = l.readIdent()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		}

		tok = newToken(token.ILLEGAL, l.ch)
	}

	l.readChar()
	return tok
}

func (l *lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *lexer) skipComment() {
	for l.ch != '\n' && l.ch != '\r' {
		l.readChar()
	}
	l.skipWhitespace()
}

func (l *lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *lexer) readTwoCharToken(tokenType token.Type) token.Token {
	ch := l.ch
	l.readChar()
	return token.Token{
		Type:    tokenType,
		Literal: string(ch) + string(l.ch),
	}
}

func (l *lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *lexer) read(checkFn func(byte) bool) string {
	position := l.position
	for checkFn(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *lexer) readIdent() string {
	return l.read(isLetter)
}

func (l *lexer) readNumber() string {
	return l.read(isDigit)
}

func (l *lexer) readNumberToken() token.Token {
	intPart := l.readNumber()
	if l.ch != '.' {
		return token.Token{
			Type:    token.INT,
			Literal: intPart,
		}
	}

	l.readChar()
	fracPart := l.readNumber()
	return token.Token{
		Type:    token.FLOAT,
		Literal: intPart + "." + fracPart,
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType token.Type, ch byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
	}
}
