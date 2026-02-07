package evaluator

import (
	"cogen/ast"
	"cogen/object" // Assuming this is the path to your object package
	"cogen/token"
	"fmt"
	"strings"
)

// ConvertSExprToAST now takes the inner Value of an object.List
func ConvertSExprToAST(input []object.Object) (*ast.Program, error) {
	prog := &ast.Program{}
	if len(input) == 0 {
		return prog, nil
	}

	// Input structure: ((('ackerman-2 n) (label1 ...) (label2 ...)))
	// The first element is the root list containing header and labels
	rootList, ok := input[0].(*object.List)
	if !ok {
		return nil, fmt.Errorf("expected list at root, got %s", input[0].Type())
	}
	root := rootList.Value

	// 1. Parse Header: ('ackerman-2 n)
	headerList, ok := root[0].(*object.List)
	if !ok {
		return nil, fmt.Errorf("expected header list, got %s", root[0].Type())
	}
	header := headerList.Value

	// Get program name from the first symbol
	if nameSym, ok := header[0].(object.ValueString); ok {
		// Trim the prefix here
		prog.Name = strings.TrimPrefix(nameSym.GetValue(), "'")
	}

	for i := 1; i < len(header); i++ {
		if varSym, ok := header[i].(object.ValueString); ok {
			varName := varSym.GetValue()
			prog.Variables = append(prog.Variables, ast.Input{
				Ident: &ast.Identifier{
					Token: token.Token{Type: token.IDENT, Literal: varName},
					Value: varName,
				},
			})
		}
	}

	// 2. Parse Label Statements: ('ack0-2 (return ('3)))
	for i := len(root)-1; i > 0; i-- {
		labelBlockObj, ok := root[i].(*object.List)
		if !ok {
			continue
		}
		labelBlock := labelBlockObj.Value

		labelVal := ""
		if lSym, ok := labelBlock[0].(object.ValueString); ok {
			// Trim the prefix here too
			labelVal = strings.TrimPrefix(lSym.GetValue(), "'")
		}

		labelStmt := &ast.LabelStatement{
			Token: token.Token{Type: token.LABEL, Literal: labelVal},
			Label: ast.Label{Token: token.Token{Type: token.LABEL, Literal: labelVal}, Value: labelVal},
		}

		for j := 1; j < len(labelBlock); j++ {
			stmtList, ok := labelBlock[j].(*object.List)
			if !ok {
				return nil, fmt.Errorf("statement must be a list, got %s", labelBlock[j].Type())
			}

			stmt, err := parseStatement(stmtList.Value)
			if err != nil {
				return prog, err
			}
			if stmt != nil {
				labelStmt.Statements = append(labelStmt.Statements, stmt)
			}
		}
		prog.Statements = append(prog.Statements, labelStmt)
	}

	return prog, nil
}

func parseStatement(list []object.Object) (ast.Statement, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("empty statement list")
	}

	opSym, ok := list[0].(object.ValueString)
	if !ok {
		return nil, fmt.Errorf("expected operator symbol, got %s", list[0].Type())
	}
	op := opSym.GetValue()

	switch op {
	case "return":
		val, err := parseExpression(list[1])
		if err != nil {
			return nil, err
		}
		return &ast.ReturnStatement{
			Token:       token.Token{Type: token.RETURN, Literal: "return"},
			ReturnValue: val,
		}, nil

	case "if":
		val, err := parseExpression(list[1])
		if err != nil {
			return nil, err
		}
		return &ast.IfStatement{
			Token:      token.Token{Type: token.IF, Literal: "if"},
			Cond:       val,
			LabelTrue:  parseTargetLabel(list[2]),
			LabelFalse: parseTargetLabel(list[3]),
		}, nil

	default:
		// Check for assignment: (n := ...)
		if len(list) >= 3 {
			midSym, ok := list[1].(object.ValueString)
			if ok && midSym.GetValue() == ":=" {
				val, err := parseExpression(list[2])
				if err != nil {
					return nil, err
				}
				identStr := op // 'op' is the first element
				return &ast.AssignmentStatement{
					Left:  &ast.Identifier{Token: token.Token{Type: token.IDENT, Literal: identStr}, Value: identStr},
					Token: token.Token{Type: token.ASSIGN, Literal: ":="},
					Right: val,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("unknown operator %s", op)
}

func parseExpression(expr object.Object) (ast.Expression, error) {
	switch v := expr.(type) {
	case *object.Integer:
		sVal := fmt.Sprint(v.Value)
		return &ast.IntegerLiteral{
			Token: token.Token{Type: token.NUMBER, Literal: sVal},
			Value: v.Value,
		}, nil

	case *object.Symbol:
		return &ast.Identifier{
			Token: token.Token{Type: token.IDENT, Literal: v.Value},
			Value: v.Value,
		}, nil

	case *object.List:
		list := v.Value
		if len(list) == 0 {
			return nil, fmt.Errorf("empty expression list")
		}

		// Helper to safely get the raw value of a symbol or string
		getRaw := func(obj object.Object) string {
			if s, ok := obj.(object.ValueString); ok {
				return s.GetValue()
			}
			return ""
		}

		first := getRaw(list[0])

		// 1. Handle (quote 3) or (' 3) -> Length 2
		if (first == "quote" || first == "'") && len(list) == 2 {
			val, err := parseExpression(list[1])
			if err != nil {
				return nil, err
			}
			return &ast.Constant{
				Token: token.Token{Type: token.QUOTE, Literal: "'"},
				Value: val,
			}, nil
		}

		// 2. Handle (call (ack 1)) -> Length 2
		if first == "call" && len(list) == 2 {
			return &ast.CallExpression{
				Token: token.Token{Type: token.CALL, Literal: "call"},
				Label: parseTargetLabel(list[1]),
			}, nil
		}

		// 3. Handle Infix: (n + 1) or (n = 0) -> Length 3
		if len(list) == 3 {
			left, err := parseExpression(list[0])
			if err != nil {
				return nil, err
			}

			op := getRaw(list[1])
			if op == "" {
				return nil, fmt.Errorf("expected operator symbol at index 1, got %s", list[1].Type())
			}

			right, err := parseExpression(list[2])
			if err != nil {
				return nil, err
			}

			return &ast.InfixExpression{
				Left:     left,
				Operator: op,
				Right:    right,
				Token:    token.Token{Literal: op},
			}, nil
		}

		// 4. Fallback: Handle shorthand literals that might be wrapped in a list
		if len(list) == 1 && first != "" {
			return &ast.Constant{
				Token: token.Token{Type: token.QUOTE, Literal: "'"},
				Value: &ast.Identifier{Token: token.Token{Type: token.IDENT, Literal: first}, Value: first},
			}, nil
		}
	}

	return nil, fmt.Errorf("unknown expression type %s for value: %s", expr.Type(), expr.String())
}

func parseTargetLabel(input object.Object) ast.Label {
	var fullLabel string

	switch v := input.(type) {
	case object.ValueString:
		fullLabel = v.GetValue()
	case *object.List:
		var parts []string
		for _, item := range v.Value {
			if s, ok := item.(object.ValueString); ok {
				parts = append(parts, s.GetValue())
			} else {
				parts = append(parts, item.String())
			}
		}
		fullLabel = strings.Join(parts, "_")
	default:
		fullLabel = input.String()
	}

	return ast.Label{
		Token: token.Token{Type: token.LABEL, Literal: fullLabel},
		Value: fullLabel,
	}
}
