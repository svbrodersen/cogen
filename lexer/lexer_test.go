package lexer

import (
	"cogen/token"
	"testing"
)

type test struct {
	expectedType    token.TokenType
	expectedLiteral string
}

func TestNextTokenEasy(t *testing.T) {
	input := `
	read x, y;
	1: if x = y goto 7 else 2;
	2: if x < y goto 5 else 3;
	3: x := x - y;
		 goto 1;
	5: y := y - x;
		 goto 1;
	7: return x;
	`
	tests := []test{
		{token.IDENT, "read"},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "1"},
		{token.COLON, ":"},
		{token.IF, "if"},
		{token.IDENT, "x"},
		{token.EQUAL, "="},
		{token.IDENT, "y"},
		{token.GOTO, "goto"},
		{token.NUMBER, "7"},
		{token.ELSE, "else"},
		{token.NUMBER, "2"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "2"},
		{token.COLON, ":"},
		{token.IF, "if"},
		{token.IDENT, "x"},
		{token.LESSTHAN, "<"},
		{token.IDENT, "y"},
		{token.GOTO, "goto"},
		{token.NUMBER, "5"},
		{token.ELSE, "else"},
		{token.NUMBER, "3"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "3"},
		{token.COLON, ":"},
		{token.IDENT, "x"},
		{token.ASSIGN, ":="},
		{token.IDENT, "x"},
		{token.SUB, "-"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.GOTO, "goto"},
		{token.NUMBER, "1"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "5"},
		{token.COLON, ":"},
		{token.IDENT, "y"},
		{token.ASSIGN, ":="},
		{token.IDENT, "y"},
		{token.SUB, "-"},
		{token.IDENT, "x"},
		{token.SEMICOLON, ";"},
		{token.GOTO, "goto"},
		{token.NUMBER, "1"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "7"},
		{token.COLON, ":"},
		{token.RETURN, "return"},
		{token.IDENT, "x"},
		{token.SEMICOLON, ";"},
	}
	l := New(input)
	testEquality(l, tests, t)
}

func TestNextTokenAck(t *testing.T) {
	input := `
	ack (m, n);
	1: if m = 0 goto done else next;
	next: if n = 0 goto ack0 else ack1;
	done: return n + 1;
	ack0: n:= 1;
				goto ack2;
	ack1: n := n - 1;
				n := call ack m n;
				goto ack2;
	ack2: m := m - 1;
				n := call ack m n;
				return n;
	`
	tests := []test{
		{token.IDENT, "ack"},
		{token.LPAREN, "("},
		{token.IDENT, "m"},
		{token.COMMA, ","},
		{token.IDENT, "n"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "1"},
		{token.COLON, ":"},
		{token.IF, "if"},
		{token.IDENT, "m"},
		{token.EQUAL, "="},
		{token.NUMBER, "0"},
		{token.GOTO, "goto"},
		{token.IDENT, "done"},
		{token.ELSE, "else"},
		{token.IDENT, "next"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "next"},
		{token.COLON, ":"},
		{token.IF, "if"},
		{token.IDENT, "n"},
		{token.EQUAL, "="},
		{token.NUMBER, "0"},
		{token.GOTO, "goto"},
		{token.IDENT, "ack0"},
		{token.ELSE, "else"},
		{token.IDENT, "ack1"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "done"},
		{token.COLON, ":"},
		{token.RETURN, "return"},
		{token.IDENT, "n"},
		{token.ADD, "+"},
		{token.NUMBER, "1"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "ack0"},
		{token.COLON, ":"},
		{token.IDENT, "n"},
		{token.ASSIGN, ":="},
		{token.NUMBER, "1"},
		{token.SEMICOLON, ";"},
		{token.GOTO, "goto"},
		{token.IDENT, "ack2"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "ack1"},
		{token.COLON, ":"},
		{token.IDENT, "n"},
		{token.ASSIGN, ":="},
		{token.IDENT, "n"},
		{token.SUB, "-"},
		{token.NUMBER, "1"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "n"},
		{token.ASSIGN, ":="},
		{token.CALL, "call"},
		{token.IDENT, "ack"},
		{token.IDENT, "m"},
		{token.IDENT, "n"},
		{token.SEMICOLON, ";"},
		{token.GOTO, "goto"},
		{token.IDENT, "ack2"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "ack2"},
		{token.COLON, ":"},
		{token.IDENT, "m"},
		{token.ASSIGN, ":="},
		{token.IDENT, "m"},
		{token.SUB, "-"},
		{token.NUMBER, "1"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "n"},
		{token.ASSIGN, ":="},
		{token.CALL, "call"},
		{token.IDENT, "ack"},
		{token.IDENT, "m"},
		{token.IDENT, "n"},
		{token.SEMICOLON, ";"},
		{token.RETURN, "return"},
		{token.IDENT, "n"},
		{token.SEMICOLON, ";"},
	}
	l := New(input)
	testEquality(l, tests, t)
}

func TestLexerIdent(t *testing.T) {
	input := `foobar;`

	l := New(input)
	tests := []test{
		{token.IDENT, "foobar"},
	}
	testEquality(l, tests, t)
}

func TestNotEqual(t *testing.T) {
	input := "!=;"
	l := New(input)
	tests := []test{
		{token.NOT_EQUAL, "!="},
		{token.SEMICOLON, ";"},
	}
	testEquality(l, tests, t)
}

func testEquality(l Lexer, tests []test, t *testing.T) {
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
