package object

import "fmt"

type ObjectType int

const (
	INTEGER ObjectType = iota
	BOOLEAN
	SYMBOL
	NULL
	RETURN_VALUE
	ERROR
)

func (ot ObjectType) String() string {
	names := [...]string{"INTEGER", "BOOLEAN", "SYMBOL", "NULL", "RETURN VALUE", "ERROR"}
	if int(ot) < 0 || int(ot) >= len(names) {
		return fmt.Sprintf("ObjectType(%d)", ot)
	}
	return names[ot]
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

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN }

type Symbol struct {
	Value string
}

func (s *Symbol) Inspect() string {
	if len(s.Value) == 1 {
		return fmt.Sprintf("'%s", s.Value)
	} else {
		return fmt.Sprintf("'(%s)", s.Value)
	}
}
func (s *Symbol) Type() ObjectType { return SYMBOL }

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
