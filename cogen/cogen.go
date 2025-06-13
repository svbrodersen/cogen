package cogen

import (
	"cogen/ast"
	"cogen/parser"
	"cogen/token"
	"errors"
	"fmt"
	"log"
	"maps"
	"sort"
	"strconv"
)

type State struct {
	delta        map[string]*ast.Identifier
	extension    *ast.Program
	curStatement *ast.LabelStatement
}

type Cogen struct {
	state    *State
	origProg *ast.Program
	parser   *parser.Parser
}

func New(p *parser.Parser) *Cogen {
	return &Cogen{
		parser: p,
	}
}

func (c *Cogen) Gen(delta []int) *ast.Program {
	c.origProg = c.parser.ParseProgram()
	// Note the first var as static
	c.state = &State{}
	c.state.delta = make(map[string]*ast.Identifier)
	for _, i := range delta {
		c.addDelta(c.origProg.Variables[i])
	}
	c.state.extension = &ast.Program{}

	c.processHeader(c.origProg)

	for _, stmt := range c.origProg.Statements {
		c.processPoly(stmt)
	}

	return c.state.extension
}

func (c *Cogen) saveState() *State {
	delta := make(map[string]*ast.Identifier)
	maps.Copy(delta, c.state.delta)
	return &State{
		delta:        delta,
		extension:    c.state.extension,
		curStatement: c.state.curStatement,
	}
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
		value = value + " " + x
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
	switch v := exp.(type) {
	case *ast.CallExpression:
		log.Fatal("exprtup: how is this call expression?")
	case *ast.ArbitraryExpression:
		log.Fatal("exprup: got arbitrary expression")
	case *ast.Identifier:
		if c.existsDelta(v) {
			return &ast.ArbitraryExpression{
				Token: newToken(token.IDENT, "list"),
				Value: "'quote " + v.Value,
			}
		} else {
			return &ast.Constant{
				Token: newToken(token.CONSTANT, "'"),
				Value: v,
			}
		}
	case ast.Value:
		return &ast.Constant{
			Token: newToken(token.CONSTANT, "'"),
			Value: v,
		}
	case *ast.InfixExpression:
		v.Left = c.exprUplift(v.Left)
		v.Right = c.exprUplift(v.Right)
		return &ast.ArbitraryExpression{
			Token: newToken(token.IDENT, "list"),
			Value: v.String(),
		}
	case *ast.PrefixExpression:
		v.Right = c.exprUplift(v.Right)
		return &ast.ArbitraryExpression{
			Token: newToken(token.IDENT, "list"),
			Value: v.String(),
		}
	default:
		log.Fatalf("exprup: reached default via: %T", v)
	}
	return nil
}

func (c *Cogen) processBlock(stmt *ast.LabelStatement) *ast.LabelStatement {
	l := c.newLabel(4, stmt.Label.Value)
	if c.existsLabel(&l.Label) {
		l, err := c.getCurLabelStatement(&l.Label)
		if err != nil {
			log.Fatalf("block: %v", err)
		}
		return l
	}

	// add new label, and attempt to processBody
	c.state.extension.Statements = append(c.state.extension.Statements, l)
	c.state.curStatement = l
	c.processBody(stmt.Statements)
	return c.state.curStatement
}

func (c *Cogen) processPoly(stmt *ast.LabelStatement) *ast.LabelStatement {
	// If already exists, then just return same state
	l1 := c.newLabel(1, stmt.Label.Value)
	if c.existsLabel(&l1.Label) {
		l, err := c.getCurLabelStatement(&l1.Label)
		if err != nil {
			log.Fatalf("block: %v", err)
		}
		return l
	}

	// Otherwise we must be able to fill this new LabelStatement

	upliftL := c.labelUplift(stmt.Label.Value)

	l3 := c.newLabel(3, stmt.Label.Value)
	l4 := c.newLabel(4, stmt.Label.Value)
	l1.Statements = []ast.Statement{
		&ast.IfStatement{
			Token: newToken(token.IF, "if"),
			Cond: &ast.ArbitraryExpression{
				Token: token.Token{
					Type:    token.IDENT,
					Literal: "done?",
				},
				Value: "(" + upliftL.String() + ")" + " code",
			},
			LabelTrue: ast.Label{
				Token: newToken(token.IDENT, "2"),
				Value: "2",
			},
			LabelFalse: l3.Label,
		},
	}

	l3.Statements = []ast.Statement{
		&ast.AssignmentStatement{
			Left:  newIdentifier("code"),
			Token: newToken(token.ASSIGN, ":="),
			Right: &ast.ArbitraryExpression{
				Token: newToken(token.IDENT, "newblock"),
				Value: "code " + "(" + upliftL.String() + ")",
			},
		},
		&ast.GotoStatement{
			Token: token.Token{
				Type:    token.GOTO,
				Literal: "goto",
			},
			Label: l4.Label,
		},
	}
	c.state.extension.Statements = append(c.state.extension.Statements, l1, l3)
	cur_state := c.saveState()
	c.processBlock(stmt)
	c.state = cur_state
	return l1
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
		log.Fatalf("expected jump, got %s", v.String())
	}
}

// Adds the block to the end of the state.extension.Statements
// Returns the label statement just added
func (c *Cogen) copyBlock(stmt *ast.LabelStatement) *ast.Label {
	newStmt := c.newLabel(5, stmt.Label.Value)
	newStmt.Statements = make([]ast.Statement, len(stmt.Statements))

	prev, _ := c.getCurLabelStatement(&newStmt.Label)
	if prev != nil {
		return &prev.Label
	}

	// Otherwise, copy it over
	for i, item := range stmt.Statements {
		switch v := item.(type) {
		case *ast.IfStatement:
			ifStmt := ast.IfStatement{
				Token:      v.Token,
				Cond:       v.Cond,
				LabelTrue:  *c.copyBlocks(&v.LabelTrue),
				LabelFalse: *c.copyBlocks(&v.LabelFalse),
			}
			newStmt.Statements[i] = &ifStmt

		case *ast.GotoStatement:
			gotoStmt := ast.GotoStatement{
				Token: v.Token,
				Label: *c.copyBlocks(&v.Label),
			}
			newStmt.Statements[i] = &gotoStmt

		default:
			newStmt.Statements[i] = v
		}
	}

	c.state.extension.Statements = append(c.state.extension.Statements, newStmt)
	return &newStmt.Label
}

func (c *Cogen) copyBlocks(label *ast.Label) *ast.Label {
	for _, stmt := range c.origProg.Statements {
		if stmt.Label.Value == label.Value {
			l := c.copyBlock(stmt)
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
			Left:  newIdentifier(stmt.Left.Value),
			Token: newToken(token.ASSIGN, ":="),
			Right: stmt.Right,
		})
	} else {
		upliftE := c.exprUplift(stmt.Left)
		c.addStatement(codeAssign(
			&ast.ArbitraryExpression{
				Token: newToken(token.IDENT, "o"),
				Value: "code " + underlineAssign(stmt.Left, upliftE),
			}))
	}
}

func (c *Cogen) processCallAssginment(stmt *ast.AssignmentStatement, callExp *ast.CallExpression) {
	// live exp
	if c.isSubsetDelta(callExp.Variables) {
		c.addDelta(stmt.Left)
		c.addStatement(
			&ast.AssignmentStatement{
				Left:  stmt.Left,
				Token: stmt.Token,
				Right: &ast.CallExpression{
					Token:     newToken(token.CALL, "call"),
					Label:     *c.copyBlocks(&callExp.Label),
					Variables: []*ast.Identifier{},
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

				Variables: []*ast.Identifier{},
			}))
		c.addStatement(codeAssign(
			&ast.ArbitraryExpression{
				Token: newToken(token.IDENT, "o"),
				Value: underlineCall(stmt.Left, upliftL.String()),
			}))

		callStmt, err := c.getOrigLabelStatement(&callExp.Label)
		if err != nil {
			log.Fatalf("call assignment: %v", err)
		}
		cur_state := c.saveState()
		c.processPoly(callStmt)
		c.state = cur_state
	}
}

func (c *Cogen) getOrigLabelStatement(stmt *ast.Label) (*ast.LabelStatement, error) {
	for _, ogStmt := range c.origProg.Statements {
		if stmt.String() == ogStmt.Label.String() {
			return ogStmt, nil
		}
	}
	msg := fmt.Sprintf("unable to find label %s", stmt.String())
	return nil, errors.New(msg)
}

func (c *Cogen) getCurLabelStatement(stmt *ast.Label) (*ast.LabelStatement, error) {
	for _, curStmt := range c.state.extension.Statements {
		if stmt.Value == curStmt.Label.Value {
			return curStmt, nil
		}
	}
	msg := fmt.Sprintf("unable to find label %s", stmt.String())
	return nil, errors.New(msg)
}

func (c *Cogen) processIf(stmt *ast.IfStatement) {
	variables := getVars(stmt.Cond)
	if c.isSubsetDelta(variables) {
		cur_state := c.saveState()
		// process true label statement
		subStmt, err := c.getOrigLabelStatement(&stmt.LabelTrue)
		if err != nil {
			log.Fatalf("if statement: %v", err)
		}
		l1 := c.processBlock(subStmt)

		c.state = cur_state

		// process false label statement
		subStmt, err = c.getOrigLabelStatement(&stmt.LabelFalse)
		if err != nil {
			log.Fatalf("if statement: %v", err)
		}
		l2 := c.processBlock(subStmt)

		// Reset to the current block, and add the if statement
		c.state = cur_state
		c.addStatement(&ast.IfStatement{
			Token:      stmt.Token,
			Cond:       stmt.Cond,
			LabelTrue:  l1.Label,
			LabelFalse: l2.Label,
		})
	} else {
		lu1 := c.labelUplift(stmt.LabelTrue.String())
		lu2 := c.labelUplift(stmt.LabelFalse.String())
		eu := c.exprUplift(stmt.Cond)

		cur_state := c.saveState()
		l1, err := c.getOrigLabelStatement(&stmt.LabelTrue)
		if err != nil {
			log.Fatalf("if statement: %v", err)
		}
		l1 = c.processPoly(l1)
		c.state = cur_state

		l2, err := c.getOrigLabelStatement(&stmt.LabelTrue)
		if err != nil {
			log.Fatalf("if statement: %v", err)
		}
		l2 = c.processPoly(l2)

		c.state = cur_state
		c.addStatement(
			codeAssign(&ast.CallExpression{
				Token: token.Token{
					Type:    token.CALL,
					Literal: "call",
				},
				Variables: []*ast.Identifier{},
				Label:     l1.Label,
			}))

		c.addStatement(
			codeAssign(&ast.CallExpression{
				Token: token.Token{
					Type:    token.CALL,
					Literal: "call",
				},
				Label:     l2.Label,
				Variables: []*ast.Identifier{},
			}))

		c.addStatement(&ast.ReturnStatement{
			Token: token.Token{
				Type:    token.RETURN,
				Literal: "return",
			},
			ReturnValue: &ast.ArbitraryExpression{
				Token: token.Token{
					Type:    token.IDENT,
					Literal: "o",
				},
				Value: "code" + underlineIf(eu, lu1.String(), lu2.String()),
			},
		})

	}
}

func (c *Cogen) processReturn(stmt *ast.ReturnStatement) {
	eu := c.exprUplift(stmt.ReturnValue)
	c.addStatement(&ast.ReturnStatement{
		Token: stmt.Token,
		ReturnValue: &ast.ArbitraryExpression{
			Token: token.Token{
				Type:    token.IDENT,
				Literal: "o",
			},
			Value: "code " + underlineReturn(eu),
		},
	})
}

func (c *Cogen) processGoto(stmt *ast.GotoStatement) {
	og, err := c.getOrigLabelStatement(&stmt.Label)
	if err != nil {
		log.Fatalf("goto: %v", err)
	}
	c.processBody(og.Statements)
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
	_, found := c.state.delta[item.Value]
	log.Printf("exists delta: %s, item: %s result: %v", c.state.delta, item.Value, found)
	log.Println("Current label statement:\n", c.state.curStatement)
	return found
}

func (c *Cogen) existsLabel(label *ast.Label) bool {
	for _, item := range c.state.extension.Statements {
		if item.Label.Value == label.Value {
			return true
		}
	}
	return false
}

func (c *Cogen) addDelta(item *ast.Identifier) {
	log.Println("add delta: ", item)
	c.state.delta[item.Value] = item
}

func (c *Cogen) removeDelta(item *ast.Identifier) {
	delete(c.state.delta, item.Value)
}

func (c *Cogen) newLabel(num int, label string) *ast.LabelStatement {
	l := strconv.Itoa(num) + "-" + label
	items := make([]string, 0, len(c.state.delta))
	for item := range c.state.delta {
		items = append(items, item)
	}
	// We have to sort to not duplicate after add and delete
	sort.Strings(items)
	for _, item := range items {
		l = l + "-" + item
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

func underlineAssign(x *ast.Identifier, exp ast.Expression) string {
	return "(list '" + x.Value + "':=" + exp.String() + ")"
}

func underlineIf(e ast.Expression, l1 string, l2 string) string {
	return "(list 'if " + e.String() + " " + l1 + " " + l2 + ")"
}

func underlineReturn(e ast.Expression) string {
	return "(list 'return " + e.String() + ")"
}

func getVars(exp ast.Expression) []*ast.Identifier {
	keys := []*ast.Identifier{}
	vars := make(map[*ast.Identifier]struct{})

	switch v := exp.(type) {
	case *ast.CallExpression:
		log.Fatal("how is this call expression?")
	case *ast.ArbitraryExpression:
		log.Fatal("getVars: got arbitrary expression")
	case *ast.PrefixExpression:
		keys = append(keys, getVars(v.Right)...)
	case *ast.InfixExpression:
		keys = append(keys, getVars(v.Left)...)
		keys = append(keys, getVars(v.Right)...)
	case *ast.List:
		for _, item := range v.Value {
			keys = append(keys, getVars(item)...)
		}
	case *ast.Constant:
		keys = append(keys, getVars(v.Value)...)
	case *ast.Identifier:
		vars[v] = struct{}{}
	}

	for i := range vars {
		keys = append(keys, i)
	}
	return keys
}
