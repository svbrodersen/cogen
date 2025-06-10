package cogen

import (
	"cogen/ast"
	"cogen/token"
	"log"
	"strconv"
)

type State struct {
	delta        map[*ast.Identifier]struct{}
	extension    ast.Program
	curStatement *ast.LabelStatement
}

type Cogen struct {
	state    *State
	origProg *ast.Program
}

func New() *Cogen {
	return &Cogen{}
}

func (c *Cogen) Gen(prog ast.Program) ast.Program {
	// Note the first var as static
	c.state = &State{}
	c.origProg = &prog
	c.state.delta = make(map[*ast.Identifier]struct{})
	for i := range prog.Variables[:len(prog.Variables)-1] {
		c.addDelta(prog.Variables[i])
	}
	c.state.extension = ast.Program{}

	c.processHeader(&prog)

	for _, stmt := range prog.Statements {
		c.processPoly(stmt)
	}

	return c.state.extension
}

func (c *Cogen) processHeader(prog *ast.Program) {
	dynamicVar := prog.Variables[len(prog.Variables)-1]
	upliftL := c.labelUplift("0")

	initLabel := c.newLabel(0, "0")
	gotoLabel := c.newLabel(1, "0")
	codeIdentifier := newIdentifier("code")

	initLabel.Statements = []ast.Statement{
		&ast.AssignmentStatement{
			Left: codeIdentifier,
			Token: token.Token{
				Type:    token.ASSIGN,
				Literal: ":=",
			},
			Right: &ast.ArbitraryExpression{
				Token: token.Token{
					Type:    token.IDENT,
					Literal: "newheader",
				},
				Value: "'(" + dynamicVar.String() + ")" + " " + upliftL.String(),
			},
		},
		&ast.GotoStatement{
			Token: token.Token{
				Type:    token.GOTO,
				Literal: "goto",
			},
			Label: gotoLabel.Label,
		},
	}

	ltoken := token.Token{
		Type:    token.LABEL,
		Literal: "2",
	}
	twoLabel := &ast.LabelStatement{
		Token: ltoken,
		Label: ast.Label{
			Token: ltoken,
			Value: ltoken.Literal,
		},
		Statements: []ast.Statement{
			&ast.ReturnStatement{
				Token: token.Token{
					Type:    token.RETURN,
					Literal: "return",
				},
				ReturnValue: newIdentifier("code"),
			},
		},
	}

	c.state.extension.Statements = append(c.state.extension.Statements, initLabel, twoLabel)
}

func (c *Cogen) labelUplift(label string) ast.Expression {
	value := "'" + label

	for x := range c.state.delta {
		value = value + " " + x.String()
	}
	stmt := &ast.ArbitraryExpression{
		Token: token.Token{
			Type:    token.IDENT,
			Literal: "list",
		},
		Value: value,
	}
	return stmt
}

func (c *Cogen) exprUplift(exp ast.Expression) ast.Expression {
}

func (c *Cogen) processBlock(stmt *ast.LabelStatement) {
	l := c.newLabel(4, stmt.TokenLiteral())
	if c.existsLabel(&l.Label) {
		return
	}

	// add new label, and attempt to processBody
	c.state.extension.Statements = append(c.state.extension.Statements, l)
	c.state.curStatement = l
	c.processBody(stmt.Statements)
}

func (c *Cogen) processPoly(stmt *ast.LabelStatement) {
	// If already exists, then just return same state
	l := c.newLabel(1, stmt.TokenLiteral())
	if c.existsLabel(&l.Label) {
		return
	}

	// Otherwise we must be able to fill this new LabelStatement

	upliftL := c.labelUplift(stmt.Label.Value)

	l1 := c.newLabel(1, l.TokenLiteral())
	l3 := c.newLabel(3, l.TokenLiteral())
	l1.Statements = []ast.Statement{
		&ast.IfStatement{
			Token: newToken(token.IF, "if"),
			Cond: &ast.ArbitraryExpression{
				Token: token.Token{
					Type:    token.IDENT,
					Literal: "done?",
				},
				Value: "l' code",
			},
			LabelTrue: ast.Label{
				Token: newToken(token.IDENT, "2"),
				Value: "2",
			},
			LabelFalse: ast.Label{
				Token: newToken(token.IDENT, l3.TokenLiteral()),
				Value: l3.TokenLiteral(),
			},
		},
	}

	l3.Statements = []ast.Statement{
		&ast.AssignmentStatement{
			Left:  newIdentifier("code"),
			Token: newToken(token.ASSIGN, ":="),
			Right: &ast.ArbitraryExpression{
				Token: newToken(token.IDENT, "newblock"),
				Value: "code " + upliftL.TokenLiteral(),
			},
		},
	}
	c.state.extension.Statements = append(c.state.extension.Statements, l1, l3)
	c.processBlock(stmt)
}

func (c *Cogen) processBody(stmts []ast.Statement) {
	for i, stmt := range stmts {
		switch v := stmt.(type) {
		case *ast.AssignmentStatement:
			c.processAssginment(v)

		default:
			c.processJump(v)
			if i != len(stmts)-1 {
				log.Fatalf("expected last statement to be jump, got %T", v)
			}
		}
	}
}

func (c *Cogen) processJump(stmt ast.Statement) {
	switch v := stmt.(type) {
	case *ast.IfStatement:
		c.processIf(v)

	case *ast.ReturnStatement:
		c.processReturn(v)
	case *ast.GotoStatement:
		c.processGoto(v)

	default:
		log.Fatalf("expected jump, got %T", v)
	}
}

// Adds the block to the end of the state.extension.Statements
// Returns the label statement just added
func (c *Cogen) copyBlock(stmt ast.LabelStatement) *ast.LabelStatement {
	value := "5-" + stmt.Label.Value
	stmt.Label = ast.Label{
		Token: newToken(token.LABEL, value),
		Value: value,
	}

	for i, item := range stmt.Statements {
		// If we have either a goto or if, then we have to copy the blocks pointed
		// to by them aswell
		switch v := item.(type) {
		case *ast.IfStatement:
			lT := c.copyBlocks(&v.LabelTrue)
			lF := c.copyBlocks(&v.LabelTrue)
			v.LabelTrue = lT.Label
			v.LabelFalse = lF.Label
			stmt.Statements[i] = v

		case *ast.GotoStatement:
			v.Label = c.copyBlocks(&v.Label).Label
			stmt.Statements[i] = v
		}
	}

	// Now all statements in this block have been cleaned, so we can add the stmt
	c.state.extension.Statements = append(c.state.extension.Statements, &stmt)
	return &stmt
}

func (c *Cogen) copyBlocks(label *ast.Label) *ast.LabelStatement {
	for _, stmt := range c.origProg.Statements {
		if stmt.Label.Value == label.Value {
			l := c.copyBlock(*stmt)
			return l
		}
	}
	log.Fatalf("expected to have a label %s, got none", label.Value)
	// not reached
	return nil
}

func (c *Cogen) addStatement(stmt ast.Statement) {
	c.state.curStatement.Statements = append(c.state.curStatement.Statements, stmt)
}

func (c *Cogen) processAssginment(stmt *ast.AssignmentStatement) {
	callExp, ok := stmt.Right.(*ast.CallExpression)
	if ok {
		c.processCallAssginment(stmt, callExp)
	} else {
		c.processRegularAssginment(stmt)
	}
}

func (c *Cogen) processRegularAssginment(stmt *ast.AssignmentStatement) {
	vars := getVars(stmt.Right)

	if c.isSubsetDelta(vars) {
		c.addDelta(stmt.Left)
		c.addStatement(&ast.AssignmentStatement{
			Left:  newIdentifier("x"),
			Token: newToken(token.ASSIGN, ":="),
			Right: stmt.Right,
		})
	} else {
		upliftE := c.exprUplift(stmt.Left)
		c.addStatement(codeAssign(
			&ast.ArbitraryExpression{
				Token: newToken(token.IDENT, "o"),
				Value: underlineAssign(stmt.Left, upliftE.String()),
			}))
	}
}

func (c *Cogen) processCallAssginment(stmt *ast.AssignmentStatement, callExp *ast.CallExpression) {
	// live exp
	if c.isSubsetDelta(callExp.Variables) {
		c.addDelta(stmt.Left)
		l := c.copyBlocks(&callExp.Label)
		c.addStatement(
			&ast.AssignmentStatement{
				Left:  stmt.Left,
				Token: stmt.Token,
				Right: &ast.CallExpression{
					Token: newToken(token.CALL, "call"),
					Label: l.Label,
				},
			},
		)
	} else {
		upliftL := c.labelUplift(callExp.Label.Value)
		c.removeDelta(stmt.Left)
		l1 := c.newLabel(1, callExp.Label.Value)
		c.addStatement(
			codeAssign(&ast.CallExpression{
				Token: newToken(token.CALL, "call"),
				Label: l1.Label,
			}))
		c.addStatement(codeAssign(
			&ast.ArbitraryExpression{
				Token: newToken(token.IDENT, "o"),
				Value: underlineCall(stmt.Left, upliftL.String()),
			}))

		// we have to save the current statement, as poly might change it
		curStatement := c.state.curStatement
		for _, ogStmt := range c.origProg.Statements {
			if callExp.Label == ogStmt.Label {
				c.processPoly(ogStmt)
				break
			}
			log.Fatalf("call: unable to find label %s", &callExp.Label)
		}
		c.state.curStatement = curStatement
	}
}

func (c *Cogen) processIf(stmt *ast.IfStatement) {
}

func (c *Cogen) processReturn(stmt *ast.ReturnStatement) {
}

func (c *Cogen) processGoto(stmt *ast.GotoStatement) {
}

func (c *Cogen) isSubsetDelta(vars []*ast.Identifier) bool {
	for _, value := range vars {
		if !c.existsDelta(value) {
			return false
		}
	}
	return true
}

func (c *Cogen) existsDelta(item *ast.Identifier) bool {
	_, found := c.state.delta[item]
	return found
}

func (c *Cogen) existsLabel(label *ast.Label) bool {
	for _, item := range c.state.extension.Statements {
		if item.TokenLiteral() == label.TokenLiteral() {
			return true
		}
	}
	return false
}

func (c *Cogen) addDelta(item *ast.Identifier) {
	c.state.delta[item] = struct{}{}
}

func (c *Cogen) removeDelta(item *ast.Identifier) {
	delete(c.state.delta, item)
}

func (c *Cogen) newLabel(num int, label string) *ast.LabelStatement {
	l := strconv.Itoa(num) + "-" + label
	for item := range c.state.delta {
		l = l + "-" + item.String()
	}

	token := token.Token{Type: token.LABEL, Literal: l}
	return &ast.LabelStatement{
		Token:      token,
		Label:      ast.Label{Token: token, Value: l},
		Statements: []ast.Statement{},
	}
}

func newToken(tokenType token.TokenType, val string) token.Token {
	return token.Token{Type: tokenType, Literal: val}
}

func newIdentifier(name string) *ast.Identifier {
	return &ast.Identifier{
		Token: token.Token{
			Type:    token.IDENT,
			Literal: name,
		},
		Value: name,
	}
}

func codeAssign(right ast.Expression) *ast.AssignmentStatement {
	return &ast.AssignmentStatement{
		Left:  newIdentifier("code"),
		Token: newToken(token.ASSIGN, ":="),
		Right: right,
	}
}

func underlineCall(x *ast.Identifier, l string) string {
	return "(list '" + x.Value + "':=(list 'call " + l + "))"
}

func getVars(exp ast.Expression) []*ast.Identifier {
	vars := make(map[*ast.Identifier]struct{})

	switch v := exp.(type) {
	case *ast.CallExpression:
		log.Fatal("how is this call expression?")
	case *ast.ArbitraryExpression:
		log.Fatal("getVars: got arbitrary expression")
	case *ast.PrefixExpression:
		getVars(v.Right)
	case *ast.InfixExpression:
		getVars(v.Left)
		getVars(v.Right)
	case *ast.List:
		for _, item := range v.Value {
			getVars(item)
		}
	case *ast.Constant:
		getVars(v.Value)
	case *ast.Identifier:
		vars[v] = struct{}{}
	}

	keys := make([]*ast.Identifier, 0, len(vars))
	for i := range vars {
		keys = append(keys, i)
	}
	return keys
}
