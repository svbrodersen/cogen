package ast

import (
	"bytes"
	"cogen/token"
	"strings"
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

type Input struct {
	Ident *Identifier
	Value string
}

type Program struct {
	Name       string
	Variables  []Input
	Statements []*LabelStatement
}

type LabelStatement struct {
	Token      token.Token // label token, updated in parser
	Label      Label
	Statements []Statement
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

type Identifier struct {
	Token token.Token
	Value string
}

type CallExpression struct {
	Token     token.Token // call
	Label     Label
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

type BooleanLiteral struct {
	Token token.Token //true or false
	Value bool
}

type PrimitiveCall struct {
	Token     token.Token // "(" Identifier token
	Primitive Expression  // Identifier
	Arguments []Expression
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
	Value []Expression
}

type SymbolExpression struct {
	Token token.Token // could be anything
	Value string
}

type Constant struct {
	Token token.Token // '
	Value Expression
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
	if p.Name != "" {
		out.WriteString(p.Name)
		args := []string{}
		for _, a := range p.Variables {
			args = append(args, a.Ident.String())
		}
		out.WriteString("(")
		out.WriteString(strings.Join(args, ", "))
		out.WriteString("):\n")
	}
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

func (fc *PrimitiveCall) expressionNode()      {}
func (fc *PrimitiveCall) TokenLiteral() string { return fc.Token.Literal }
func (fc *PrimitiveCall) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range fc.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(fc.Primitive.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

func (cs *Constant) expressionNode()      {}
func (cs *Constant) TokenLiteral() string { return cs.Token.Literal }
func (cs *Constant) String() string {
	var out bytes.Buffer

	out.WriteString(cs.Token.Literal)
	out.WriteString(cs.Value.String())
	return out.String()
}

func (se *SymbolExpression) expressionNode()      {}
func (se *SymbolExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SymbolExpression) String() string {
	var out bytes.Buffer
	out.WriteString(se.Value)
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
	out.WriteString(ce.Label.String())
	return out.String()
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }

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
			out.WriteString(stmt.String() + ";\n")
		} else {
			out.WriteString(stmt.String() + ";\n\t")
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
	out.WriteString(" ")
	out.WriteString(as.TokenLiteral())
	out.WriteString(" ")
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
	return out.String()
}

func (is *IfStatement) statementNode() {}
func (is *IfStatement) TokenLiteral() string {
	return is.Token.Literal
}

func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString(is.TokenLiteral() + " ")
	out.WriteString(is.Cond.String())
	out.WriteString(" ")
	out.WriteString(is.LabelTrue.String() + " ")
	out.WriteString("else ")
	out.WriteString(is.LabelFalse.String())
	return out.String()
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string {
	return i.Value
}

func (i *Identifier) Equal(o *Identifier) bool {
	return i.String() == o.String()
}

func (ll *Label) expressionNode()      {}
func (ll *Label) TokenLiteral() string { return ll.Token.Literal }
func (ll *Label) String() string       { return ll.Value }

func (ll *List) expressionNode()      {}
func (ll *List) TokenLiteral() string { return ll.Token.Literal }
func (ll *List) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	for i, elem := range ll.Value {
		var elemStr string
		if elem == nil {
			elemStr = "nil"
		} else {
			elemStr = elem.String()
		}
		if i == len(ll.Value)-1 {
			out.WriteString(elemStr)
		} else {
			out.WriteString(elemStr + " ")
		}
	}
	out.WriteString(")")
	return out.String()
}
