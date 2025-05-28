package lexer

import (
	"cogen/token"
)

type Lexer interface {
	NextToken() token.Token
}

type DefaultLexer struct {
	input    string
	position int
	ch       byte
}

func New(input string) *DefaultLexer {
	l := &DefaultLexer{input: input}
	l.position = -1
	l.readChar()
	return l
}

func (l *DefaultLexer) readChar() {
	readPosition := l.position + 1
	if readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[readPosition]
	}
	l.position = readPosition
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
		tok = newToken(token.SEMICOLON, ';')
	case '(':
		tok = newToken(token.LPAREN, '(')
	case ')':
		tok = newToken(token.RPAREN, ')')
	case ',':
		tok = newToken(token.COMMA, ',')
	case '\'':
		tok = newToken(token.QUOTE, '\'')
	case '=':
		tok = newToken(token.EQUAL, '=')
	case '<':
		tok = newToken(token.LESSTHAN, '<')
	case '>':
		tok = newToken(token.GREATERTHAN, '>')
	case '-':
		tok = newToken(token.SUB, '-')
	case '+':
		tok = newToken(token.ADD, '+')
	case '*':
		tok = newToken(token.ASTERISK, '*')
	case '/':
		tok = newToken(token.SLASH, '/')
	case ':':
		// We have to first take care of the : := situation
		if l.peakChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.ASSIGN, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.COLON, ':')
		}
	case '!':
		// We have to first take care of the : := situation
		if l.peakChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQUAL, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.BANG, '!')
		}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF

	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.NUMBER
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
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

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
