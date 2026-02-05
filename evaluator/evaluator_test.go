package evaluator

import (
	"cogen/lexer"
	"cogen/object"
	"cogen/parser"
	"fmt"
	"os"
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
			"type mismatch: INTEGER + BOOLEAN, for: 5 true",
		},
		{
			"1: 5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN, for: 5 true",
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
			"identifier not found: foobar at 1:3",
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
		{"1: c := 1; b := call 2; return c; 2: c:= 5; return 10;", 1},
		{"1: b := call 2; b; \n2: a := 5*5; return a;", 25},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestTuringMachine(t *testing.T) {
	input, err := os.ReadFile("../turing_machine.fcl")
	if err != nil {
		t.Fatal("Failed to read file")
	}
	fmt.Printf("Result: %s\n", testEval(string(input)))
}

func TestCogenAckermann(t *testing.T) {
	input, err := os.ReadFile("../cogen_ackermann.fcl")
	if err != nil {
		t.Fatal("Failed to read file")
	}
	env := object.NewEnvironment()
	env.Set("m", &object.Integer{Value: 2})
	res := testEvalWithEnv(string(input), env)
	fmt.Printf("Result: %s\n", res)
}


func TestPrimitiveCalls(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		// hd
		{"1: hd(list(10, 20, 30));", int64(10)},
		{"1: hd('(10 20 30));", int64(10)},
		// tl
		{"1: tl(list(10, 20, 30));", []int64{20, 30}},
		{"1: tl('(10 20 30));", []int64{20, 30}},
		// list
		{"1: list(7, 8, 9);", []int64{7, 8, 9}},
		{"1: newTail(2, list('(1 1), '(2 3)));", []string{"'(2 3)"}},
		// hd error
		{"1: hd(list());", "hd called on empty list"},
		// tl error
		{"1: tl(list());", "tl called on empty list"},
		// new_tail error
		{"1: newTail(1, 2);", "newTail expects second element to be a list, got INTEGER"},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int64:
			testIntegerObject(t, evaluated, expected)
		case []int64:
			listObj, ok := evaluated.(*object.List)
			if !ok {
				t.Errorf("object is not List. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if len(listObj.Value) != len(expected) {
				t.Errorf("list has wrong length. got=%d, want=%d", len(listObj.Value), len(expected))
				continue
			}
			for i, v := range expected {
				intObj, ok := listObj.Value[i].(*object.Integer)
				if !ok || intObj.Value != v {
					t.Errorf("list element %d wrong. got=%v, want=%d", i, listObj.Value[i], v)
				}
			}
		case []string:
			listObj, ok := evaluated.(*object.List)
			if !ok {
				t.Errorf("object is not List. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if len(listObj.Value) != len(expected) {
				t.Errorf("list has wrong length. got=%d, want=%d", len(listObj.Value), len(expected))
				continue
			}
			for i, v := range expected {
				symObj, ok := listObj.Value[i].(*object.List)
				if !ok || symObj.String() != v {
					t.Errorf("list element %d wrong. got=%v, want=%s", i, listObj.Value[i], v)
				}
			}
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("expected error object. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
			}
		}
	}
}

func testEvalWithEnv(input string, env *object.Environment) object.Object {
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	evaluator := New(program)
	return evaluator.Eval(program, env)
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
