package lexer

import (
	"cogen/token"
)

type Lexer interface {
	NextToken() token.Token
	GetLine() int
	GetColumn() int
	GetInput() string
}

type DefaultLexer struct {
	input    string
	position int
	ch       byte
	line     int
	column   int
}

func (l *DefaultLexer) GetInput() string {
	return l.input
}

func New(input string) *DefaultLexer {
	l := &DefaultLexer{input: input, line: 1, column: 0}
	l.position = -1
	l.readChar()
	return l
}

func (l *DefaultLexer) GetLine() int {
	return l.line
}

func (l *DefaultLexer) GetColumn() int {
	return l.line
}

func (l *DefaultLexer) readChar() {
	readPosition := l.position + 1
	if readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[readPosition]
	}
	l.position = readPosition

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

func (l *DefaultLexer) peakChar() byte {
	if l.position+1 >= len(l.input) {
		return 0
	} else {
		return l.input[l.position+1]
	}
}

func (l *DefaultLexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()
	switch l.ch {
	case ';':
		tok = newToken(l, token.SEMICOLON, ';')
	case '(':
		tok = newToken(l, token.LPAREN, '(')
	case ')':
		tok = newToken(l, token.RPAREN, ')')
	case ',':
		tok = newToken(l, token.COMMA, ',')
	case '\'':
		tok = newToken(l, token.QUOTE, '\'')
	case '=':
		tok = newToken(l, token.EQUAL, '=')
	case '<':
		tok = newToken(l, token.LESSTHAN, '<')
	case '>':
		tok = newToken(l, token.GREATERTHAN, '>')
	case '-':
		tok = newToken(l, token.SUB, '-')
	case '+':
		tok = newToken(l, token.ADD, '+')
	case '*':
		tok = newToken(l, token.ASTERISK, '*')
	case '/':
		tok = newToken(l, token.SLASH, '/')
	case ':':
		// We have to first take care of the : := situation
		if l.peakChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = newToken(l, token.ASSIGN, string(ch)+string(l.ch))
		} else {
			tok = newToken(l, token.COLON, ':')
		}
	case '!':
		// We have to first take care of the : := situation
		if l.peakChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = newToken(l, token.NOT_EQUAL, string(ch)+string(l.ch))
		} else {
			tok = newToken(l, token.BANG, '!')
		}
	case 0:
		tok = newToken(l, token.EOF, "")
	default:
		if isLetter(l.ch) {
			literal := l.readIdentifier()
			tok = newToken(l, token.LookupIdent(literal), literal)
			return tok
		} else if isDigit(l.ch) {
			num := l.readNumber()
			tok = newToken(l, token.NUMBER, num)
			return tok
		} else {
			tok = newToken(l, token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return tok
}

func (l *DefaultLexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *DefaultLexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *DefaultLexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func newToken[T rune | string | byte](l Lexer, tokenType token.TokenType, ch T) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Line: l.GetLine(),
		Column: l.GetColumn()}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
