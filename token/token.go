package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{
	"goto":   GOTO,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"call":   CALL,
	"true":   TRUE,
	"false":  FALSE,
}

const (
	// special
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// call
	CALL = "call"

	// int
	NUMBER = "NUMBER"

	// assignment
	ASSIGN = ":="

	// bool
	TRUE  = "true"
	FALSE = "false"

	// Jump
	GOTO   = "goto"
	IF     = "if"
	RETURN = "return"
	ELSE   = "else"

	// Expr
	IDENT    = "IDENT"
	CONSTANT = "CONSTANT"
	LABEL    = "LABEL"

	// SPECIAL FOR SYMBOL EXPRESSION
	SYMBOL = "SYMBOL"

	// Operators
	EQUAL       = "="
	NOT_EQUAL   = "!="
	LESSTHAN    = "<"
	GREATERTHAN = ">"
	SUB         = "-"
	ADD         = "+"
	ASTERISK    = "*"
	SLASH       = "/"
	BANG        = "!"

	// Parenthesis
	LPAREN = "("
	RPAREN = ")"

	// Delimeters
	SEMICOLON   = ";"
	COLON       = ":"
	COMMA       = ","
	QUOTE       = "'"
	DOUBLEQUOTE = "\""
)

func LookupIdent(ident string) TokenType {
	tok, ok := keywords[ident]
	if ok {
		return tok
	}
	return IDENT
}
