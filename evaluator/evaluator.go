package evaluator

import (
	"cogen/ast"
	"cogen/object"
	"fmt"
	"log"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

type Evaluator struct {
	Program *ast.Program
}

func New(program *ast.Program) *Evaluator {
	return &Evaluator{Program: program}
}

func (e *Evaluator) Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.LabelStatement:
		return e.evalStatements(node.Statements, env)
	case *ast.Program:
		return e.evalProgram(node, env)
	case *ast.SymbolExpression:
		return &object.Symbol{Value: node.Value}
	case *ast.Constant:
		return e.Eval(node.Value, env)
	case *ast.ExpressionStatement:
		return e.Eval(node.Expression, env)
	case *ast.PrefixExpression:
		right := e.Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := e.Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := e.Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfStatement:
		return e.evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := e.Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.AssignmentStatement:
		val := e.Eval(node.Right, env)
		if isError(val) {
			return val
		}
		env.Set(node.Left.Value, val)
	case *ast.CallExpression:
		return e.evalCallExpression(node, env)
	case *ast.FunctionCall:
		return e.evalFunctionCall(node, env)
	}
	return nil
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR
	}
	return false
}

func (e *Evaluator) evalProgram(prog *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, lstms := range prog.Statements {
		result = e.evalStatements(lstms.Statements, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func (e *Evaluator) evalStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = e.Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE || rt == object.ERROR {
				return result
			}
		}
	}
	return result
}

func newError(format string, a ...any) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func (e *Evaluator) evalCallExpression(node *ast.CallExpression, env *object.Environment) object.Object {
	var labelStmt *ast.LabelStatement
	for _, v := range e.Program.Statements {
		if v.Label == node.Label {
			labelStmt = v
			break
		}
	}

	newEnv := object.NewEnclosedEnvironment(env)
	evaluated := e.Eval(labelStmt, newEnv)
	return unwrapReturnValue(evaluated)
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func (e *Evaluator) evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, exp := range exps {
		evaluated := e.Eval(exp, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func (e *Evaluator) evalFunctionCall(node *ast.FunctionCall, env *object.Environment) object.Object {
	args := e.evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	log.Printf("node: %v", node.Value)
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: %s", node.Value)
	}
	return val
}

func (e *Evaluator) evalIfExpression(stmt *ast.IfStatement, env *object.Environment) object.Object {
	condition := e.Eval(stmt.Cond, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return e.Eval(&stmt.LabelTrue, env)
	} else {
		return e.Eval(&stmt.LabelFalse, env)
	}
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s for %s", operator, right.Inspect())
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	value, ok := right.(*object.Integer)
	if !ok {
		return newError("unknown operator: -%s", right.Type())
	}
	return &object.Integer{Value: -value.Value}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func intergetInfix(operator string, leftObj, rightObj object.Object) object.Object {
	left := leftObj.(*object.Integer).Value
	right := rightObj.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: left + right}
	case "-":
		return &object.Integer{Value: left - right}
	case "*":
		return &object.Integer{Value: left * right}
	case "/":
		return &object.Integer{Value: left / right}
	case "=":
		return &object.Boolean{Value: left == right}
	case "!=":
		return &object.Boolean{Value: left != right}
	case "<":
		return &object.Boolean{Value: left < right}
	case ">":
		return &object.Boolean{Value: left > right}
	default:
		return newError("unknown operator: %s %s %s", leftObj.Type(), operator, rightObj.Type())
	}
}

func booleanInfix(operator string, leftObj, rightObj object.Object) object.Object {
	left := leftObj.(*object.Boolean).Value
	right := rightObj.(*object.Boolean).Value
	switch operator {
	case "=":
		return &object.Boolean{Value: left == right}
	case "!=":
		return &object.Boolean{Value: left != right}
	default:
		return newError("unknown operator: %s %s %s", leftObj.Type(), operator, rightObj.Type())
	}
}

func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	if left.Type() != right.Type() {
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	}
	switch left.Type() {
	case object.INTEGER:
		val := intergetInfix(operator, left, right)
		return val
	case object.BOOLEAN:
		return booleanInfix(operator, left, right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}
