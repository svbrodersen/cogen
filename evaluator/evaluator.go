package evaluator

import (
	"bytes"
	"cogen/ast"
	"cogen/object"
	"fmt"
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
	if node == nil {
		return NULL
	}
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
	case *ast.GotoStatement:
		return e.evalGotoExpression(node, env)
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
	case *ast.Label:
		return e.evalLabel(node, env)
	case *ast.List:
		return e.evalList(node, env)
	case *ast.PrimitiveCall:
		return e.evalPrimitiveCall(node, env)
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
	result := e.Eval(prog.Statements[0], env)
	switch result := result.(type) {
	case *object.ReturnValue:
		return result.Value
	case *object.Error:
		return result
	}
	return result
}

func (e *Evaluator) evalLabel(node *ast.Label, env *object.Environment) object.Object {
	var labelStmt *ast.LabelStatement
	for _, v := range e.Program.Statements {
		if v.Label.Value == node.Value {
			labelStmt = v
			break
		}
	}

	if labelStmt == nil {
		return newError("label not found: %s", node.Value)
	}

	evaluated := e.Eval(labelStmt, env)
	return evaluated
}

func (e *Evaluator) evalGotoExpression(node *ast.GotoStatement, env *object.Environment) object.Object {
	evaluated := e.Eval(&node.Label, env)
	return evaluated
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
		if v.Label.Value == node.Label.Value {
			labelStmt = v
			break
		}
	}

	if labelStmt == nil {
		return newError("LabelStatement not found in call expression: %s", node.Label.Value)
	}

	newEnv := object.NewEnclosedEnvironment(env)
	evaluated := e.Eval(labelStmt, newEnv)
	return unwrapReturnValue(evaluated)
}

func (e *Evaluator) evalList(node *ast.List, env *object.Environment) object.Object {
	value := e.evalExpressions(node.Value, env)
	return &object.List{Value: value}
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

func (e *Evaluator) evalPrimitiveCall(node *ast.PrimitiveCall, env *object.Environment) object.Object {
	args := e.evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
	var out bytes.Buffer
	for _, val := range args {
		out.WriteString(fmt.Sprintf("(%s, %s) ", val.String(), val.Type()))
	}

	return CallPrimitive(node.Primitive.String(), args)
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: %s at %d:%d", node.Value, node.Token.Line, node.Token.Column)
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
		return newError("unknown operator: %s for %s", operator, right.String())
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
		if left == right {
			return TRUE
		} else {
			return FALSE
		}
	case "!=":
		if left != right {
			return TRUE
		} else {
			return FALSE
		}
	case "<":
		if left < right {
			return TRUE
		} else {
			return FALSE
		}
	case ">":
		if left > right {
			return TRUE
		} else {
			return FALSE
		}
	default:
		return newError("unknown operator: %s %s %s", leftObj.Type(), operator, rightObj.Type())
	}
}

func booleanInfix(operator string, leftObj, rightObj object.Object) object.Object {
	left := leftObj.(*object.Boolean).Value
	right := rightObj.(*object.Boolean).Value
	switch operator {
	case "=":
		if left == right {
			return TRUE
		} else {
			return FALSE
		}
	case "!=":
		if left != right {
			return TRUE
		} else {
			return FALSE
		}
	default:
		return newError("unknown operator: %s %s %s", leftObj.Type(), operator, rightObj.Type())
	}
}

func listInfix(operator string, leftObj, rightObj object.Object) object.Object {
	left := leftObj.(*object.List).Value
	right := rightObj.(*object.List).Value
	switch operator {
	case "=":
		if len(left) != len(right) {
			return FALSE
		}
		initial := true
		for i := range left {
			initial = initial && isTruthy(evalInfixExpression(operator, left[i], right[i]))
		}
		if initial {
			return TRUE
		} else {
			return FALSE
		}
	case "!=":
		// This is just the reverse of the previous
		if len(left) != len(right) {
			return TRUE
		}
		initial := true
		for i := range left {
			initial = initial && isTruthy(evalInfixExpression("=", left[i], right[i]))
		}
		if initial {
			return FALSE
		} else {
			return TRUE
		}
	default:
		return newError("unknown operator: %s %s %s", leftObj.Type(), operator, rightObj.Type())
	}
}

func symbolInfix(operator string, leftObj object.Object, rightObj object.Object) object.Object {
	left := leftObj.(*object.Symbol).Value
	right := rightObj.(*object.Symbol).Value

	switch operator {
	case "=":
		if left == right {
			return TRUE
		}	else {
			return FALSE
		}
	case "!=":
		if left != right {
			return TRUE
		} else {
			return FALSE
		}
	default:
		return newError("unknown operator: %s %s %s", leftObj.Type(), operator, rightObj.String())
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
	case object.LIST:
		return listInfix(operator, left, right)
	case object.SYMBOL:
		return symbolInfix(operator, left, right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}
