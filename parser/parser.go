package parser

import (
	"cogen/ast"
	"cogen/lexer"
	"cogen/token"
	"fmt"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
)

var precedence = map[token.TokenType]int{
	token.EQUAL:       EQUALS,
	token.NOT_EQUAL:   EQUALS,
	token.LESSTHAN:    LESSGREATER,
	token.GREATERTHAN: LESSGREATER,
	token.ADD:         SUM,
	token.SUB:         SUM,
	token.ASTERISK:    PRODUCT,
	token.SLASH:       PRODUCT,
}

type (
	valueParseFn  func() ast.Value
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l lexer.Lexer

	errors []string

	curToken  token.Token
	peakToken token.Token

	valueParseFns  map[token.TokenType]valueParseFn
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.valueParseFns = make(map[token.TokenType]valueParseFn)
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	// value
	p.registerValue(token.IDENT, p.parseIdentifier)
	p.registerValue(token.NUMBER, p.parseIntegerLiteral)
	p.registerValue(token.LPAREN, p.parseList)
	p.registerValue(token.QUOTE, p.parseConstant)

	// prefix
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.SUB, p.parsePrefixExpression)
	p.registerPrefix(token.HD, p.parsePrefixExpression)
	p.registerPrefix(token.TL, p.parsePrefixExpression)
	p.registerPrefix(token.CONS, p.parsePrefixExpression)
	// infix
	p.registerInfix(token.ADD, p.parseInfixExpression)
	p.registerInfix(token.SUB, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQUAL, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQUAL, p.parseInfixExpression)
	p.registerInfix(token.LESSTHAN, p.parseInfixExpression)
	p.registerInfix(token.GREATERTHAN, p.parseInfixExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerValue(tokenType token.TokenType, fn valueParseFn) {
	p.valueParseFns[tokenType] = fn
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peakToken
	p.peakToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peakTokenIs(t token.TokenType) bool {
	return p.peakToken.Type == t
}

func (p *Parser) requirePeak(t token.TokenType) bool {
	if p.peakTokenIs(t) {
		p.nextToken()
		return true
	} else {
		// note the fail
		p.peakError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peakError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s", t, p.peakToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peakPrecedence() int {
	if p, ok := precedence[p.peakToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedence[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// check if we have reached a Label. Not just a simple token
func (p *Parser) peakLabel() bool {
	// if not an identifier or number, then it is not a label
	if p.curToken.Type != token.IDENT && p.curToken.Type != token.NUMBER {
		return false
	}
	// if it now has a colon afterwards, then it is a label.
	if p.peakTokenIs(token.COLON) {
		return true
	}
	return false
}

func (p *Parser) parseCall() *ast.CallExpression {
	stmt := &ast.CallExpression{Token: p.curToken}
	stmt.Variables = []*ast.Identifier{}
	p.nextToken()
	stmt.Label = p.parseLabel()
	p.nextToken()
	for !p.peakTokenIs(token.SEMICOLON) && !p.peakTokenIs(token.EOF) {
		stmt.Variables = append(stmt.Variables, p.requireIdentifier())
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseLabel() ast.Label {
	p.curToken.Type = token.LABEL
	return ast.Label{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseList() ast.Value {
	stmt := &ast.List{Token: p.curToken}
	p.nextToken()
	var values []ast.Value
	// loop as long as we don't have the closing of the list
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		value := p.parseValueExpression()
		if value == nil {
			msg := fmt.Sprintf("list: could not parse %q as integer", p.curToken.Literal)
			p.errors = append(p.errors, msg)
		}
		values = append(values, p.parseValueExpression())
		p.nextToken()

		// If we are in the middle of the list, then we remove the comma
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	stmt.Value = values
	return stmt
}

func (p *Parser) parseFunction() (string, []*ast.Identifier) {
	name := ""
	variables := []*ast.Identifier{}
	if !p.peakTokenIs(token.LPAREN) {
		return name, variables
	}
	name = p.curToken.Literal
	p.nextToken()
	// now on (
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		p.nextToken()
		variables = append(variables, p.requireIdentifier())
		p.nextToken()
	}
	// eat )
	p.nextToken()

	// In the case we reach here eat ; too
	p.nextToken()
	return name, variables
}

func (p *Parser) parseConstant() ast.Value {
	stmt := &ast.Constant{Token: p.curToken}
	p.nextToken()

	stmt.Value = p.parseValueExpression()
	return stmt
}

func (p *Parser) requireIdentifier() *ast.Identifier {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIdentifier() ast.Value {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Value {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.GOTO:
		return p.parseGotoStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.IF:
		return p.parseIfStatement()
	default:
		// If next token is assign, then take care of it
		if p.peakTokenIs(token.ASSIGN) {
			return p.parseAssginmentStatement()
		}
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLabelStatement() *ast.LabelStatement {
	// Update the label to state label
	label := p.parseLabel()
	// Move past the colon
	if !p.requirePeak(token.COLON) {
		p.nextToken()
		return nil
	}
	p.nextToken()
	labelStmt := &ast.LabelStatement{Token: p.curToken, Label: label}
	labelStmt.Statements = []ast.Statement{}

	for !p.peakLabel() && p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			labelStmt.Statements = append(labelStmt.Statements, stmt)
		}
		p.nextToken()
	}
	return labelStmt
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}
	p.nextToken()

	stmt.Cond = p.parseExpression(LOWEST)
	p.nextToken()

	// Skip over goto
	if !p.curTokenIs(token.GOTO) {
		msg := fmt.Sprintf("expected goto, got %s", p.curToken)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()
	// Parse true label
	stmt.LabelTrue = p.parseLabel()
	p.nextToken()
	// skip over else
	if !p.curTokenIs(token.ELSE) {
		msg := fmt.Sprintf("expected else, got %s", p.curToken)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()

	// parse false label
	stmt.LabelFalse = p.parseLabel()
	p.nextToken()

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)
	p.nextToken()
	return stmt
}

func (p *Parser) parseGotoStatement() *ast.GotoStatement {
	stmt := &ast.GotoStatement{Token: p.curToken}

	p.nextToken()
	stmt.Label = p.parseLabel()
	p.nextToken()
	return stmt
}

func (p *Parser) parseAssginmentStatement() *ast.AssignmentStatement {
	val := p.requireIdentifier()
	stmt := &ast.AssignmentStatement{Left: val}
	if !p.requirePeak(token.ASSIGN) {
		return nil
	}
	stmt.Token = p.curToken
	// Move over the :=
	p.nextToken()
	if p.curToken.Type == token.CALL {
		stmt.Right = p.parseCall()
	} else {
		stmt.Right = p.parseExpression(LOWEST)
	}
	p.nextToken()

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)
	p.nextToken()
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	// First attempt to parse ValuepExpression. Otherwise parse prefix
	var leftExp ast.Expression
	value := p.parseValueExpression()
	if value == nil {
		prefix := p.prefixParseFns[p.curToken.Type]
		if prefix == nil {
			p.noPrefixParseFnError(p.curToken.Type)
			return nil
		}
		leftExp = prefix()
	} else {
		leftExp = value
	}
	for !p.peakTokenIs(token.SEMICOLON) && precedence < p.peakPrecedence() {
		infix := p.infixParseFns[p.peakToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parseValueExpression() ast.Value {
	value := p.valueParseFns[p.curToken.Type]
	if value == nil {
		return nil
	}
	return value()
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)

	if !p.requirePeak(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []*ast.LabelStatement{}
	program.Name, program.Variables = p.parseFunction()

	for p.curToken.Type != token.EOF {
		stmt := p.parseLabelStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	return program
}
