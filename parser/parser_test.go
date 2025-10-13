package parser

import (
	"cogen/ast"
	"cogen/lexer"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestGotoStatements(t *testing.T) {
	input := `
	2: goto 5;
	2: goto 3;
	2: goto 1;
	2: goto ack0;
	`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if err := checkParserErrors(p); err != nil {
		t.Fatal(err)
	}

	if len(program.Statements) != 4 {
		t.Fatalf("program.Statements does not have 4 statements. Got: %d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"5"},
		{"3"},
		{"1"},
		{"ack0"},
	}

	for i, tt := range tests {
		stmt := testLabelStatement(t, program.Statements[i], "2", 1)[0]
		testGotoStatement(t, stmt, tt.expectedIdentifier)
	}
}

func testGotoStatement(t *testing.T, stmt ast.Statement, expected string) {
	if stmt.TokenLiteral() != "goto" {
		t.Fatalf("stmt.TokenLiteral not 'goto'. got=%q", stmt.TokenLiteral())
	}

	gotoStmt, ok := stmt.(*ast.GotoStatement)
	if !ok {
		t.Fatalf("stmt no *ast.GotoStatement. got %T", stmt)
	}

	if gotoStmt.Label.TokenLiteral() != expected {
		t.Fatalf("stmt.Label.TokenLiteral() not '%s'. got %s", expected, gotoStmt.Label.TokenLiteral())
	}
}

func TestReturnStatement(t *testing.T) {
	input := `
		ack: return 5;
		ack: return 10;
	`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatal("program is nil")
	}

	if err := checkParserErrors(p); err != nil {
		t.Fatal(err)
	}

	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements does not contain 2 statements, got %d", len(program.Statements))
	}

	for _, ps := range program.Statements {
		stmt := testLabelStatement(t, ps, "ack", 1)[0]
		returnstmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.ReturnStatement, got %T", stmt)
			continue
		}
		if returnstmt.TokenLiteral() != "return" {
			t.Errorf("stmt.TokenLiteral() not 'return', got %q", returnstmt.TokenLiteral())
		}

	}
}

func TestIdentifierExpression(t *testing.T) {
	input := `
		1: foobar;
	`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()

	if err := checkParserErrors(p); err != nil {
		t.Fatal(err)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement, got %d", len(program.Statements))
	}
	stmt := testLabelStatement(t, program.Statements[0], "1", 1)[0]

	exp, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got %T", stmt)
	}

	ident, ok := exp.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("expression is not *ast.Identifier, got %T", exp.Expression)
	}

	if ident.TokenLiteral() != "foobar" {
		t.Fatalf("ident.TokenLiteral() not 'foobar' got %s", ident.Value)
	}

	if ident.Value != "foobar" {
		t.Fatalf("ident.Value not 'foobar' got %s", ident.Value)
	}
}

func TestIntegerExpression(t *testing.T) {
	input := "1: 5;"
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	if err := checkParserErrors(p); err != nil {
		t.Fatal(err)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement, got %d", len(program.Statements))
	}

	stmt := testLabelStatement(t, program.Statements[0], "1", 1)[0]
	exp, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got %T", stmt)
	}

	literal, ok := exp.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got %T", stmt)
	}

	if literal.Value != 5 {
		t.Errorf("literal.Value is not 5, got %d", &literal.Value)
	}

	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.Value is not 5, got %s", literal.TokenLiteral())
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"1: -a * b;",
			"1: ((-a) * b);\n",
		},
		{
			"1: !-a;",
			"1: (!(-a));\n",
		},
		{
			"1: a + b + c;",
			"1: ((a + b) + c);\n",
		},
		{
			"1: a + b - c;",
			"1: ((a + b) - c);\n",
		},
		{
			"1: a * b * c;",
			"1: ((a * b) * c);\n",
		},
		{
			"1: a * b / c;",
			"1: ((a * b) / c);\n",
		},
		{
			"1: a + b / c;",
			"1: (a + (b / c));\n",
		},
		{
			"1: a + b * c + d / e - f;",
			"1: (((a + (b * c)) + (d / e)) - f);\n",
		},
		{
			`1: 3 + 4;
				-5 * 5;`,
			"1: (3 + 4);\n\t((-5) * 5);\n",
		},
		{
			"1: 5 > 4 = 3 < 4;",
			"1: ((5 > 4) = (3 < 4));\n",
		},
		{
			"1: 5 < 4 != 3 > 4;",
			"1: ((5 < 4) != (3 > 4));\n",
		},
		{
			"1: 3 + 4 * 5 = 3 * 1 + 4 * 5;",
			"1: ((3 + (4 * 5)) = ((3 * 1) + (4 * 5)));\n",
		},
		{
			"1: 3 + 4 * 5 = 3 * 1 + 4 * 5;",
			"1: ((3 + (4 * 5)) = ((3 * 1) + (4 * 5)));\n",
		},
	}
	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		if err := checkParserErrors(p); err != nil {
			t.Fatalf("error in tests[%d]: %v", i, err)
		}
		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestLabel(t *testing.T) {
	input := `
		1: 5 + 5;
		3: 6 * 6;
	`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatal("program is nil")
	}

	if err := checkParserErrors(p); err != nil {
		t.Fatal(err)
	}

	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements does not contain 2 statements, got %d", len(program.Statements))
	}
	stmt := testLabelStatement(t, program.Statements[0], "1", 1)[0]

	exp, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("ls.Statement[0] not ExpressionStatement, got %T", program.Statements[3])
	}
	testInfixExpression(t, exp.Expression, 5, "+", 5)

	stmt = testLabelStatement(t, program.Statements[1], "3", 1)[0]

	exp, ok = stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("ls.Statement[1] not ExpressionStatement, got %T", program.Statements[3])
	}
	testInfixExpression(t, exp.Expression, 6, "*", 6)
}

func TestConstantSimple(t *testing.T) {
	input := `
		1: x := '1;
	`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatal("program is nil")
	}

	if err := checkParserErrors(p); err != nil {
		t.Fatal(err)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements, got %d", len(program.Statements))
	}
	stmt := testLabelStatement(t, program.Statements[0], "1", 1)[0]

	ass, ok := stmt.(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("stmt not AssignmentStatement, got %T", stmt)
	}
	testAssignmentStatement(t, "x", ass, "'1")
}

func TestConstantHard(t *testing.T) {
	input := `
		1: x := '1;
			 goto 2;
		2: y := '(1 + 3 := 3);
		3: y := (3 + 1);
	`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatal("program is nil")
	}

	if err := checkParserErrors(p); err != nil {
		t.Fatal(err)
	}

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements, got %d", len(program.Statements))
	}

	stmts := testLabelStatement(t, program.Statements[0], "1", 2)
	testAssignmentStatement(t, "x", stmts[0], "'1")
	testGotoStatement(t, stmts[1], "2")

	stmt := testLabelStatement(t, program.Statements[1], "2", 1)[0]
	testAssignmentStatement(t, "y", stmt, "'(1 + 3 := 3)")

	stmt = testLabelStatement(t, program.Statements[2], "3", 1)[0]
	testAssignmentStatement(t, "y", stmt, "(3 + 1)")
}

func TestAckermannFunc(t *testing.T) {
	input := `
		ack (m, n);
		1: if m = 0 goto done else next;
		next: if n = 0 goto ack0 else ack1;
		done: return n + 1;
		ack0: n := 1;
					goto ack2;
		ack1: n := n - 1;
					n := call ack m n;
					goto ack2;
		ack2: m := m - 1;
			n := call ack m n;
			return n;
	`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatal("program is nil")
	}

	if err := checkParserErrors(p); err != nil {
		t.Fatal(err)
	}

	if len(program.Statements) != 6 {
		t.Fatalf("program.Statements does not contain 6 statements, got %d", len(program.Statements))
	}

	if program.Name != "ack" {
		t.Fatalf("program.Name is not 'ack', got %s", program.Name)
	}

	testIdentifier(t, program.Variables[0].Ident, "m")
	testIdentifier(t, program.Variables[1].Ident, "n")

	// Check label statements
	stmt := testLabelStatement(t, program.Statements[0], "1", 1)[0]
	testIfStatement(t, stmt, "(m = 0)", "done", "next")

	stmt = testLabelStatement(t, program.Statements[1], "next", 1)[0]
	testIfStatement(t, stmt, "(n = 0)", "ack0", "ack1")

	stmt = testLabelStatement(t, program.Statements[2], "done", 1)[0]
	rs, ok := stmt.(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected return statement, got %t", stmt)
	}
	if rs.ReturnValue.String() != "(n + 1)" {
		t.Fatalf("expected (n+1), got %s", rs.ReturnValue.String())
	}
}

func checkParserErrors(p *Parser) error {
	if len(p.Errors()) == 0 {
		return nil
	}
	msg := p.GetErrorMessage()
	return errors.New(msg)
}

func testLabelStatement(t *testing.T, st ast.Statement, value string, length int) []ast.Statement {
	ls, ok := st.(*ast.LabelStatement)
	if !ok {
		t.Fatalf("program.Statement[0] not label statement, got %T", st)
	}
	if ls.Label.Value != value {
		t.Fatalf("ls.Label.Value not 1, got %s", ls.Label.Value)
	}
	if len(ls.Statements) != length {
		t.Fatalf("label.Statement has more than 1 statement, got %d", len(ls.Statements))
	}
	return ls.Statements
}

func testIfStatement(t *testing.T, st ast.Statement, cond string, l1 string, l2 string) {
	is, ok := st.(*ast.IfStatement)
	if !ok {
		t.Fatalf("statement not If statement, got %T", st)
	}

	if is.Cond.String() != cond {
		t.Fatalf("if condition not '%s', got %s", cond, is.Cond.String())
	}

	if is.LabelTrue.String() != l1 {
		t.Fatalf("if true label not '%s', got %s", l1, is.LabelTrue.String())
	}
	if is.LabelFalse.String() != l2 {
		t.Fatalf("if false label not '%s', got %s", l2, is.LabelFalse.String())
	}
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not ast.Identifier, got %T", exp)
	}

	if ident.Value != value {
		t.Fatalf("ident.Value not %s. got=%s", value, ident.Value)
	}

	if ident.TokenLiteral() != value {
		t.Fatalf("ident.TokenLiteral not %s. got=%s", value, ident.TokenLiteral())
	}
}

func testAssignmentStatement(t *testing.T, left string, stmt ast.Statement, right string) {
	ass, ok := stmt.(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("stmt is not AssignmentStatement, got %T", stmt)
	}

	if ass.Left.String() != left {
		t.Fatalf("assignment.Left != %s, got %s", left, ass.Left.String())
	}

	if ass.TokenLiteral() != ":=" {
		t.Fatalf("assignment.TokenLiteral() != ':=', got %s", ass.TokenLiteral())
	}

	if ass.Right.String() != right {
		t.Fatalf("assignment.Value.String() != '%s', got %s", right, ass.Right.String())
	}
}

func testConstant(t *testing.T, exp ast.Expression, value string) {
	cons, ok := exp.(*ast.Constant)
	if !ok {
		t.Fatalf("exp not ast.ConstantStatement, got %T", exp)
	}

	if cons.TokenLiteral() != "'" {
		t.Fatalf("cons.TokenLiteral() is not ', got %T", cons.TokenLiteral())
	}

	if cons.String() == value {
		t.Fatalf("cons.Value is not %T, got %T", value, cons.Value)
	}
}

func testList(t *testing.T, exp ast.Expression, value []ast.Expression) {
	l, ok := exp.(*ast.List)
	if !ok {
		t.Fatalf("exp not ast.List, got %T", exp)
	}

	if reflect.DeepEqual(l.Value, value) {
		t.Fatalf("ident.Value not %d. got=%d", value, l.Value)
	}
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, value int64) {
	il, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not ast.IntegerLiteral, got %T", exp)
	}

	if il.Value != value {
		t.Fatalf("ident.Value not %d. got=%d", value, il.Value)
	}

	if il.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Fatalf("ident.TokenLiteral not %d. got=%s", value, il.TokenLiteral())
	}
}

func testLiteralExpression[T int | int64 | string](t *testing.T, exp ast.Expression, expected T) {
	switch v := any(expected).(type) {
	case int:
		testIntegerLiteral(t, exp, int64(v))
	case int64:
		testIntegerLiteral(t, exp, v)
	case string:
		if v[0] == '\'' {
			testConstant(t, exp, v)
		} else {
			testIdentifier(t, exp, v)
		}
	default:
		t.Fatalf("type of exp not handled, got %T", exp)
	}
}

func testInfixExpression[T int | int64 | string](t *testing.T, exp ast.Expression, left T, operator string, right T) {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("exp is not ast.InfixExpression, got %T", exp)
	}
	testLiteralExpression(t, opExp.Left, left)

	if opExp.Operator != operator {
		t.Fatalf("exp.Operator is not '%s', got %s", operator, opExp.Operator)
	}
	testLiteralExpression(t, opExp.Right, right)
}
