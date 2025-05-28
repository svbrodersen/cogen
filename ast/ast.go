package ast

import (
	"bytes"
	"cogen/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Value interface {
	Expression
	valueNode()
}

type Program struct {
	Name       string
	Variables  []*Identifier
	Statements []Statement
}

type LabelStatement struct {
	Token      token.Token // label token, updated in parser
	Label      Label
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type Identifier struct {
	Token token.Token
	Value string
}

type Label struct {
	Token token.Token
	Value string
}

type GotoStatement struct {
	Token token.Token // goto token
	Label Label
}

type ReturnStatement struct {
	Token       token.Token // return token
	ReturnValue Expression
}

type CallExpression struct {
	Token     token.Token // call
	Label     Label
	Variables []*Identifier
}

type IfStatement struct {
	Token      token.Token // if
	Cond       Expression  // expression
	LabelTrue  Label
	LabelFalse Label
}

type ExpressionStatement struct {
	Token      token.Token // initial token
	Expression Expression
}

type AssignmentStatement struct {
	Left  *Identifier
	Token token.Token // :=
	Right Expression  // Some value type
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

type PrefixExpression struct {
	Token    token.Token // prefix token e.g. hd
	Operator string
	Right    Expression
}

type InfixExpression struct {
	Token    token.Token // prefix token e.g. hd
	Left     Expression
	Operator string
	Right    Expression
}

type List struct {
	Token token.Token // (
	Value []Value
}

type Constant struct {
	Token token.Token // '
	Value Value
}

func (cs *Constant) valueNode()           {}
func (cs *Constant) expressionNode()      {}
func (cs *Constant) TokenLiteral() string { return cs.Token.Literal }
func (cs *Constant) String() string {
	var out bytes.Buffer

	out.WriteString(cs.Token.Literal)
	out.WriteString(cs.Value.String())
	return out.String()
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	// Add space between prefix if not - or !
	if pe.Token.Type != token.BANG && pe.Token.Type != token.SUB {
		out.WriteString(" ")
	}
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ce.TokenLiteral() + " ")
	out.WriteString(ce.Label.String() + " ")
	for i, ident := range ce.Variables {
		if i == len(ce.Variables)-1 {
			out.WriteString(ident.String() + ";")
		} else {
			out.WriteString(ident.String() + " ")
		}
	}
	return out.String()
}

func (i *IntegerLiteral) valueNode()            {}
func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

func (ls *LabelStatement) statementNode() {}
func (ls *LabelStatement) TokenLiteral() string {
	return ls.Token.Literal
}

func (ls *LabelStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.Label.String() + ": ")
	for i, stmt := range ls.Statements {
		if i == len(ls.Statements)-1 {
			out.WriteString(stmt.String() + ";")
		} else {
			out.WriteString(stmt.String() + ";\n  ")
		}
	}
	return out.String()
}

func (as *AssignmentStatement) statementNode() {}
func (as *AssignmentStatement) TokenLiteral() string {
	return as.Token.Literal
}

func (as *AssignmentStatement) String() string {
	var out bytes.Buffer
	out.WriteString(as.Left.String())
	out.WriteString(as.TokenLiteral())
	out.WriteString(as.Right.String())
	return out.String()
}

func (rs *ReturnStatement) statementNode() {}
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

func (gt *GotoStatement) statementNode() {}
func (gt *GotoStatement) TokenLiteral() string {
	return gt.Token.Literal
}

func (gt *GotoStatement) String() string {
	var out bytes.Buffer

	out.WriteString(gt.TokenLiteral() + " ")
	out.WriteString(gt.Label.String())
	out.WriteString(";")
	return out.String()
}

func (is *IfStatement) statementNode() {}
func (is *IfStatement) TokenLiteral() string {
	return is.Token.Literal
}

func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString(is.TokenLiteral() + " ")
	out.WriteString(is.Cond.String() + " ")
	out.WriteString(is.LabelTrue.String() + " ")
	out.WriteString(is.LabelFalse.String() + " ")
	return out.String()
}

func (i *Identifier) valueNode()      {}
func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string {
	return i.Value
}

func (ll *Label) expressionNode()      {}
func (ll *Label) TokenLiteral() string { return ll.Token.Literal }
func (ll *Label) String() string       { return ll.Value }

func (i *List) valueNode()            {}
func (ll *List) expressionNode()      {}
func (ll *List) TokenLiteral() string { return ll.Token.Literal }
func (ll *List) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	for i, elem := range ll.Value {
		if i == len(ll.Value)-1 {
			out.WriteString(elem.String())
		} else {
			out.WriteString(elem.String() + ", ")
		}
	}
	out.WriteString(")")
	return out.String()
}
