package generator

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
	state           *State
	OriginalProgram *ast.Program
	dynamicVar      []ast.Expression
	parser          *parser.Parser
}

func New(p *parser.Parser) *Cogen {
	return &Cogen{
		parser: p,
	}
}

func (c *Cogen) Gen(delta []int) (*ast.Program, error) {
	// Parse program and check for errors
	c.OriginalProgram = c.parser.ParseProgram()

	if len(c.parser.Errors()) != 0 {
		return nil, errors.New(c.parser.GetErrorMessage())
	}
	log.Println(c.OriginalProgram)
	// Note the first var as static
	c.state = &State{}
	c.state.delta = make(map[string]*ast.Identifier, len(delta))
	vars := make([]*ast.Identifier, len(delta))
	for i, delt := range delta {
		cpy := *c.OriginalProgram.Variables[delt]
		c.addDelta(&cpy)
		vars[i] = &cpy
	}
	c.state.extension = &ast.Program{
		Name:      c.OriginalProgram.Name,
		Variables: vars,
	}

	c.processHeader()

	// Start the process on the first statement
	c.processPoly(c.OriginalProgram.Statements[0])

	return c.state.extension, nil
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

func (c *Cogen) dynamicVariables() []ast.Expression {
	vars := make(
		[]ast.Expression,
		len(c.OriginalProgram.Variables)-len(c.state.extension.Variables),
	)
	count := 0
	for _, item := range c.OriginalProgram.Variables {
		if !c.existsDelta(item) {
			vars[count] = item
			count += 1
		}
	}
	return vars
}

func (c *Cogen) processHeader() {
	dynamicVar := c.dynamicVariables()
	upliftL := c.labelUplift(c.OriginalProgram.Name)
	initialLabel := c.OriginalProgram.Statements[0].Label

	initLabel := c.newLabel(0, initialLabel.Value)
	gotoLabel := c.newLabel(1, initialLabel.Value)
	codeIdentifier := newIdentifier("code")
	newHeader := newIdentifier("newheader")

	initLabel.Statements = []ast.Statement{
		&ast.AssignmentStatement{
			Left: codeIdentifier,
			Token: token.Token{
				Type:    token.ASSIGN,
				Literal: ":=",
			},
			Right: &ast.FunctionCall{
				Token:     newHeader.Token,
				Function:  newHeader,
				Arguments: append(dynamicVar, upliftL),
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

	c.state.extension.Statements = append(
		c.state.extension.Statements,
		initLabel,
		twoLabel,
	)
}

func (c *Cogen) labelUplift(label string) ast.Expression {
	listFunc := ast.FunctionCall{
		Token:    newToken(token.LPAREN, "("),
		Function: newIdentifier("list"),
	}
	arguments := make(
		[]ast.Expression,
		len(c.state.delta)+1,
	)
	arguments[0] = &ast.Constant{
		Token: newToken(token.CONSTANT, "'"),
		Value: newSymbol(label),
	}

	items := make([]*ast.Identifier, len(c.state.delta))
	counter := 0
	for _, item := range c.state.delta {
		cpy := *item
		items[counter] = &cpy
		counter += 1
	}
	// We have to sort it, such that the naming is always the same.
	sort.Slice(items[:], func(i, j int) bool {
		return items[i].Value < items[j].Value
	})

	for i, item := range items {
		// take the initial label into account
		arguments[i+1] = item
	}

	listFunc.Arguments = arguments

	stmt := listFunc
	return &stmt
}

func (c *Cogen) exprUplift(exp ast.Expression) ast.Expression {
	switch v := exp.(type) {
	case *ast.CallExpression:
		log.Fatal("exprup: how is this call expression?")
	case *ast.Identifier:
		if c.existsDelta(v) {
			return &ast.FunctionCall{
				Token:    newToken(token.LPAREN, "("),
				Function: newIdentifier("list"),
				Arguments: []ast.Expression{
					&ast.Constant{
						Token: newToken(token.CONSTANT, "'"),
						Value: newSymbol("quote"),
					},
					v,
				},
			}
		} else {
			return &ast.Constant{
				Token: newToken(token.CONSTANT, "'"),
				Value: newSymbol(v.String()),
			}
		}
	case *ast.InfixExpression:
		arguments := make([]ast.Expression, 3)
		arguments[1] = &ast.Constant{
			Token: newToken(token.CONSTANT, "'"),
			Value: newSymbol(v.Operator),
		}
		arguments[0] = c.exprUplift(v.Left)
		arguments[2] = c.exprUplift(v.Right)

		return &ast.FunctionCall{
			Token:     newToken(token.LPAREN, "("),
			Function:  newIdentifier("list"),
			Arguments: arguments,
		}
	case *ast.PrefixExpression:
		arguments := make([]ast.Expression, 2)
		arguments[0] = &ast.Constant{
			Token: newToken(token.CONSTANT, "'"),
			Value: newSymbol(v.Operator),
		}
		arguments[1] = c.exprUplift(v.Right)

		return &ast.FunctionCall{
			Token:     newToken(token.LPAREN, "("),
			Function:  newIdentifier("list"),
			Arguments: arguments,
		}
	case *ast.FunctionCall:
		newStmt := &ast.FunctionCall{
			Token:    v.Token,
			Function: v.Function,
		}
		copy(newStmt.Arguments, v.Arguments)
		return newStmt
	default:
		return v
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
	doneFunc := newIdentifier("is_done")
	code := newIdentifier("code")
	l1.Statements = []ast.Statement{
		&ast.IfStatement{
			Token: newToken(token.IF, "if"),
			Cond: &ast.FunctionCall{
				Token:     doneFunc.Token,
				Function:  doneFunc,
				Arguments: []ast.Expression{upliftL, code},
			},
			LabelTrue: ast.Label{
				Token: newToken(token.IDENT, "2"),
				Value: "2",
			},
			LabelFalse: l3.Label,
		},
	}

	newblock := newIdentifier("newblock")
	l3.Statements = []ast.Statement{
		&ast.AssignmentStatement{
			Left:  newIdentifier("code"),
			Token: newToken(token.ASSIGN, ":="),
			Right: &ast.FunctionCall{
				Token:     newblock.Token,
				Function:  newblock,
				Arguments: []ast.Expression{code, upliftL},
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
	c.processBlock(stmt)
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

	// Then we add the statement
	c.state.extension.Statements = append(c.state.extension.Statements, newStmt)

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

	return &newStmt.Label
}

func (c *Cogen) copyBlocks(label *ast.Label) *ast.Label {
	for _, stmt := range c.OriginalProgram.Statements {
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
	switch expr := stmt.Right.(type) {
	case *ast.CallExpression:
		c.processCallAssginment(stmt, expr)
	case *ast.FunctionCall:
		c.processFunctionCall(stmt, expr)
	default:
		c.processRegularAssginment(stmt)
	}
}

func (c *Cogen) processRegularAssginment(stmt *ast.AssignmentStatement) {
	vars := getVars(stmt.Right)

	if c.isSubsetDelta(vars) {
		c.addStatement(&ast.AssignmentStatement{
			Left:  newIdentifier(stmt.Left.Value),
			Token: newToken(token.ASSIGN, ":="),
			Right: stmt.Right,
		})
		c.addDelta(stmt.Left)
	} else {
		upliftE := c.exprUplift(stmt.Right)
		code := newIdentifier("code")
		o := newIdentifier("o")
		leftCpy := *stmt.Left
		c.addStatement(codeAssign(
			&ast.FunctionCall{
				Token:     o.Token,
				Function:  o,
				Arguments: []ast.Expression{code, underlineAssign(&leftCpy, upliftE)},
			}))
		c.removeDelta(stmt.Left)
	}
}

func (c *Cogen) processCallAssginment(
	stmt *ast.AssignmentStatement,
	callExp *ast.CallExpression,
) {
	// live exp
	if c.isSubsetDelta(callExp.Variables) {
		leftCpy := *stmt.Left
		c.addStatement(
			&ast.AssignmentStatement{
				Left:  &leftCpy,
				Token: stmt.Token,
				Right: &ast.CallExpression{
					Token:     newToken(token.CALL, "call"),
					Label:     *c.copyBlocks(&callExp.Label),
					Variables: []*ast.Identifier{},
				},
			},
		)
		c.addDelta(stmt.Left)
	} else {
		// first process poly on the label
		callStmt, err := c.getOrigLabelStatement(&callExp.Label)
		if err != nil {
			log.Fatalf("call assignment: %v", err)
		}
		curState := c.saveState()
		c.processPoly(callStmt)
		c.state = curState

		// then add our code
		upliftL := c.labelUplift(callExp.Label.Value)
		l1 := c.newLabel(1, callExp.Label.Value)
		c.addStatement(
			codeAssign(&ast.CallExpression{
				Token:     newToken(token.CALL, "call"),
				Label:     l1.Label,
				Variables: []*ast.Identifier{},
			}))
		code := newIdentifier("code")
		o := newIdentifier("o")
		leftCpy := *stmt.Left
		c.addStatement(codeAssign(
			&ast.FunctionCall{
				Token:     o.Token,
				Function:  o,
				Arguments: []ast.Expression{code, underlineCall(&leftCpy, upliftL)},
			}))

		// and update delta
		c.removeDelta(stmt.Left)
	}
}

func (c *Cogen) processFunctionCall(
	stmt *ast.AssignmentStatement,
	call *ast.FunctionCall,
) {
	callCpy := *call
	code := newIdentifier("code")
	o := newIdentifier("o")
	leftCpy := *stmt.Left
	c.addStatement(codeAssign(
		&ast.FunctionCall{
			Token:     o.Token,
			Function:  o,
			Arguments: []ast.Expression{code, underlineCall(&leftCpy, &callCpy)},
		}))
	// and update delta
	c.removeDelta(stmt.Left)
}

func (c *Cogen) getOrigLabelStatement(stmt *ast.Label) (*ast.LabelStatement, error) {
	for _, ogStmt := range c.OriginalProgram.Statements {
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
		newStmt := &ast.IfStatement{
			Token: stmt.Token,
			Cond:  stmt.Cond,
		}
		c.addStatement(newStmt)
		curState := c.saveState()
		// process true label statement
		subStmt, err := c.getOrigLabelStatement(&stmt.LabelTrue)
		if err != nil {
			log.Fatalf("if statement: %v", err)
		}
		l1 := c.processBlock(subStmt)

		// Reset state before we process false
		c.state = curState

		// process false label statement
		subStmt, err = c.getOrigLabelStatement(&stmt.LabelFalse)
		if err != nil {
			log.Fatalf("if statement: %v", err)
		}
		l2 := c.processBlock(subStmt)

		// Reset to the current block, and add the labels
		c.state = curState
		newStmt.LabelTrue = l1.Label
		newStmt.LabelFalse = l2.Label
	} else {
		lu1 := c.labelUplift(stmt.LabelTrue.String())
		lu2 := c.labelUplift(stmt.LabelFalse.String())
		eu := c.exprUplift(stmt.Cond)

		curState := c.saveState()
		l1, err := c.getOrigLabelStatement(&stmt.LabelTrue)
		if err != nil {
			log.Fatalf("if statement: %v", err)
		}
		l1 = c.processPoly(l1)

		c.state = curState
		curState = c.saveState()
		l2, err := c.getOrigLabelStatement(&stmt.LabelFalse)
		if err != nil {
			log.Fatalf("if statement: %v", err)
		}
		l2 = c.processPoly(l2)

		c.state = curState
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

		o := newIdentifier("o")
		code := newIdentifier("code")
		c.addStatement(&ast.ReturnStatement{
			Token: token.Token{
				Type:    token.RETURN,
				Literal: "return",
			},
			ReturnValue: &ast.FunctionCall{
				Token:     o.Token,
				Function:  o,
				Arguments: []ast.Expression{code, underlineIf(eu, lu1, lu2)},
			},
		})

	}
}

func (c *Cogen) processReturn(stmt *ast.ReturnStatement) {
	eu := c.exprUplift(stmt.ReturnValue)
	o := newIdentifier("o")
	code := newIdentifier("code")
	c.addStatement(&ast.ReturnStatement{
		Token: stmt.Token,
		ReturnValue: &ast.FunctionCall{
			Token:     o.Token,
			Function:  o,
			Arguments: []ast.Expression{code, underlineReturn(eu)},
		},
	})
}

func (c *Cogen) processGoto(stmt *ast.GotoStatement) {
	ogStmt, err := c.getOrigLabelStatement(&stmt.Label)
	if err != nil {
		log.Fatalf("goto: %v", err)
	}
	curState := c.saveState()
	c.processBody(ogStmt.Statements)
	c.state = curState
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
	itemCpy := *item
	c.state.delta[item.Value] = &itemCpy
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

func newFunction(function ast.Expression, arguments []ast.Expression) ast.Expression {
	return &ast.FunctionCall{
		Token:     newToken(token.LPAREN, "("),
		Function:  function,
		Arguments: arguments,
	}
}

func newConstant(value ast.Expression) ast.Expression {
	return &ast.Constant{
		Token: newToken(token.CONSTANT, "'"),
		Value: value,
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

func newSymbol(name string) ast.Expression {
	return &ast.SymbolExpression{
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

func underlineCall(x *ast.Identifier, l ast.Expression) ast.Expression {
	arg1 := []ast.Expression{
		newConstant(newSymbol("call")),
		l,
	}
	arguments := []ast.Expression{
		newConstant(newSymbol(x.String())),
		newConstant(newSymbol(":=")),
		newFunction(newSymbol("list"), arg1),
	}
	return newFunction(newSymbol("list"), arguments)
}

func underlineAssign(x *ast.Identifier, exp ast.Expression) ast.Expression {
	arguments := []ast.Expression{
		newConstant(newSymbol(x.String())),
		newConstant(newSymbol(":=")),
		exp,
	}
	return newFunction(newSymbol("list"), arguments)
}

func underlineIf(e ast.Expression,
	l1 ast.Expression,
	l2 ast.Expression,
) ast.Expression {
	arguments := []ast.Expression{
		newConstant(newSymbol("if")),
		e,
		l1,
		l2,
	}
	return newFunction(newIdentifier("list"), arguments)
}

func underlineReturn(e ast.Expression) ast.Expression {
	arguments := []ast.Expression{
		newConstant(newSymbol("return")),
		e,
	}
	return newFunction(newIdentifier("list"), arguments)
}

func getVars(exp ast.Expression) []*ast.Identifier {
	keys := []*ast.Identifier{}
	vars := make(map[*ast.Identifier]struct{})

	switch v := exp.(type) {
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
	case *ast.FunctionCall:
		for _, item := range v.Arguments {
			keys = append(keys, getVars(item)...)
		}
	case *ast.Identifier:
		vars[v] = struct{}{}
	default:
		return nil
	}

	for i := range vars {
		if i != nil {
			keys = append(keys, i)
		}
	}
	return keys
}
