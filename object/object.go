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
	String() string
}

type Integer struct {
	Value int64
}

func (i *Integer) String() string   { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER }
func (i *Integer) GetValue() string { return fmt.Sprint(i.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) String() string   { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN }
func (b *Boolean) GetValue() string { return fmt.Sprint(b.Value) }

type String struct {
	Value string
}

func (s *String) String() string {
	return fmt.Sprintf("%s", s.Value)
}
func (s *String) Type() ObjectType { return STRING }
func (s *String) GetValue() string { return s.Value }

type Symbol struct {
	Value string
}

func (s *Symbol) String() string {
	return s.InspectInList(false)
}

func (s *Symbol) InspectInList(inList bool) string {
	if inList {
		return s.Value
	}
	return fmt.Sprintf("'%s", s.Value)
}

func (s *Symbol) Type() ObjectType { return SYMBOL }
func (s *Symbol) GetValue() string { return s.Value }

type List struct {
	Value []Object
}

func (l *List) InspectInList(inList bool) string {
	var out bytes.Buffer
	if !inList {
		out.WriteString("'")
	}
	out.WriteString("(")
	for i, elem := range l.Value {
		var elemStr string
		if elem == nil {
			elemStr = "nil"
		} else {
			// Use InspectInList(true) if available, else fallback to Inspect()
			if s, ok := elem.(interface{ InspectInList(bool) string }); ok {
				elemStr = s.InspectInList(true)
			} else {
				elemStr = elem.String()
			}
		}
		if i == len(l.Value)-1 {
			out.WriteString(elemStr)
		} else {
			out.WriteString(elemStr + " ")
		}
	}
	out.WriteString(")")
	return out.String()
}

func (s *List) String() string {
	return s.InspectInList(false)
}
func (s *List) Type() ObjectType { return LIST }

type Null struct {
}

func (n *Null) String() string   { return "null" }
func (n *Null) Type() ObjectType { return NULL }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE }
func (rv *ReturnValue) String() string   { return rv.Value.String() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR }
func (e *Error) String() string   { return "ERROR: " + e.Message }
