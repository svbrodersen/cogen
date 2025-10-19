package object

import (
	"bytes"
	"fmt"
)

type ObjectType int

const (
	INTEGER ObjectType = iota
	BOOLEAN
	SYMBOL
	NULL
	RETURN_VALUE
	ERROR
	STRING
	LIST
)

func (ot ObjectType) String() string {
	names := [...]string{"INTEGER", "BOOLEAN", "SYMBOL", "NULL", "RETURN VALUE", "ERROR", "STRING", "LIST"}
	if int(ot) < 0 || int(ot) >= len(names) {
		return fmt.Sprintf("ObjectType(%d)", ot)
	}
	return names[ot]
}

type ValueString interface {
	GetValue() string
}

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER }
func (i *Integer) GetValue() string { return fmt.Sprint(i.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN }
func (b *Boolean) GetValue() string { return fmt.Sprint(b.Value) }

type String struct {
	Value string
}

func (s *String) Inspect() string {
	return fmt.Sprintf("%s", s.Value)
}
func (s *String) Type() ObjectType { return STRING }
func (s *String) GetValue() string { return s.Value }

type Symbol struct {
	Value string
}

func (s *Symbol) Inspect() string {
	return fmt.Sprintf("'%s", s.Value)
}
func (s *Symbol) Type() ObjectType { return SYMBOL }
func (s *Symbol) GetValue() string { return s.Value }

type List struct {
	Value []Object
}

func (s *List) Inspect() string {
	var out bytes.Buffer
	for _, v := range s.Value {
		out.WriteString(v.Inspect())
	}

	return out.String()
}
func (s *List) Type() ObjectType { return LIST }

type Null struct {
}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NULL }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
