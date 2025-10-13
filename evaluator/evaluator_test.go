package evaluator

import (
	"cogen/lexer"
	"cogen/object"
	"cogen/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1: 5;", 5},
		{"1: 10;", 10},
		{"1: -5;", -5},
		{"1: -10;", -10},
		{"1: 5 + 5 + 5 + 5 - 10;", 10},
		{"1: 2 * 2 * 2 * 2 * 2;", 32},
		{"1: -50 + 100 + -50;", 0},
		{"1: 5 * 2 + 10;", 20},
		{"1: 5 + 2 * 10;", 25},
		{"1: 20 + 2 * -10;", 0},
		{"1: 50 / 2 * 2 + 10;", 60},
		{"1: 2 * (5 + 10);", 30},
		{"1: 3 * 3 * 3 + 10;", 37},
		{"1: 3 * (3 * 3) + 10;", 37},
		{"1: (5 + 10 * 2 + 15 / 3) * 2 + -10;", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1: true;", true},
		{"5: false;", false},
		{"1: 1 < 2;", true},
		{"1: 1 > 2;", false},
		{"1: 1 < 1;", false},
		{"1: 1 > 1;", false},
		{"1: 1 = 1;", true},
		{"1: 1 != 1;", false},
		{"1: 1 = 2;", false},
		{"1: 1 != 2;", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestSymbolExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1: ':=;", ":="},
		{"2: 'test;", "test"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testSymbolObject(t, evaluated, tt.expected)
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"3: return 10;", 10},
		{"2: return 10; 9;", 10},
		{"1: return 2 * 5; 9;", 10},
		{"1: 9; return 2 * 5; 9;", 10},
		{
			`
1: return 15;
2: return 20;
`, 15}}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"1: 5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"1: 5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"1: -true",
			"unknown operator: -BOOLEAN",
		},
		{
			"1: true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"1: 5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"1: foobar;",
			"identifier not found: foobar",
		},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v), expected=%s",
				evaluated, evaluated, tt.expectedMessage)
			continue
		}
		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Message)
		}
	}

}

func TestVariables(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1: a := 5; a;", 5},
		{"1: a := 5 * 5; a;", 25},
		{"1: a := 5; b := a; b;", 5},
		{"1: a := 5; b := a; c := a + b + 5; c;", 15},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestCallExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1: a := 5; return a;\n2: b := call 1; b", 5},
		{"1: a := 5*5; return a;\n2: b := call 1; b", 25},
		{"1: a := 5*5; b:=10*10; return b;\n2: b := call 1; b", 100},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	env := object.NewEnvironment()
	evaluator := New(program)
	return evaluator.Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%v, want=%v",
			result.Value, expected)
		return false
	}
	return true
}

func testSymbolObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.Symbol)
	if !ok {
		t.Errorf("object is not Symbol. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%v, want=%v",
			result.Value, expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}
