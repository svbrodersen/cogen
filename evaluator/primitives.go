package evaluator

import (
	"cogen/object"
)

func head(arg object.Object) object.Object {
	s, ok := arg.(*object.List)
	if !ok {
		return newError("hd expects list, got %s", arg.Type())
	}
	if len(s.Value) == 0 {
		return newError("hd called on empty list")
	}
	return s.Value[0]
}

// Should be all but first element
func tail(arg object.Object) object.Object {
	s, ok := arg.(*object.List)
	if !ok {
		return newError("tl expects list, got %s", arg.Type())
	}
	if len(s.Value) == 0 {
		return newError("tl called on empty list")
	}
	return &object.List{Value: s.Value[1:]}
}

func o(s1_obj object.Object, inputs ...object.Object) object.Object {
	s1, ok := s1_obj.(*object.List)
	if !ok {
		return newError("o expected first argument to be list, got %s", s1_obj.Type())
	}
	val := append(s1.Value, inputs...)
	return &object.List{Value: val}
}

func list(a ...object.Object) object.Object {
	aCopy := make([]object.Object, len(a))
	copy(aCopy, a)
	return &object.List{Value: aCopy}
}

func cons(a object.Object, b object.Object) object.Object {
	if bLst, ok := b.(*object.List); ok {
		res := append([]object.Object{a}, bLst.Value...)
		return &object.List{Value: res}
	}

	return &object.List{Value: []object.Object{a, b}}
}

// newTail(2, '((0 if 0 goto 3) (1 right) (2 goto 0) (3 write 1)))
func newTail(item object.Object, Q_obj object.Object) object.Object {
	Q, ok := Q_obj.(*object.List)
	if !ok {
		return newError("new_tail expects second element to be a list, got %s", Q_obj.Type())
	}
	val := item.String()
	i := 0
	for _, block := range Q.Value {
		lst, ok := block.(*object.List)
		if !ok {
			return newError("new_tail expects second input to be list of list, got %s", block.Type())
		}
		if len(lst.Value) == 0 {
			continue
		}
		// We only search for symbol statements
		v, ok := lst.Value[0].(object.ValueString)
		if !ok {
			return newError("new_tail expects the first value of each sublist to implement the ValueString interface.")
		}
		if v.GetValue() == val {
			break
		}
		i++
	}
	return &object.List{Value: Q.Value[i:]}
}

func newHeader(name_obj object.Object, dynVars ...object.Object) object.Object {
	v, ok := name_obj.(*object.List)
	if !ok {
		return newError("newHeader expects first argument to be a list, got %s", name_obj.Type())
	}
	innerList := object.List{
		Value: []object.Object{},
	}

	var name string
	for i, subName := range v.Value {
		if i == len(v.Value)-1 {
			name += subName.String()
		} else {
			name += subName.String() + "-"
		}
	}
	sym := object.Symbol{
		Value: name,
	}
	innerList.Value = append(innerList.Value, &sym)

	for _, variable := range dynVars {
		innerList.Value = append(innerList.Value, variable)
	}

	header := object.List{
		Value: []object.Object{
			&innerList,
		},
	}
	return &header
}

func newBlock(code_obj object.Object, name_obj object.Object) object.Object {
	code, ok := code_obj.(*object.List)
	if !ok {
		return newError("newBlock expects first argument (code) to be a list of list, got %s", code_obj.Type())
	}
	inner_code, ok := code_obj.(*object.List)
	if !ok {
		return newError("newBlock expects first argument (code) to be a list of list, got %s", code_obj.Type())
	}

	name_list, ok := name_obj.(*object.List)
	if !ok {
		return newError("newBlock expects second argument to be a list, got %s", name_list.Type())
	}

	name := ""
	for i, subName := range name_list.Value {
		if i == len(name_list.Value)-1 {
			name += subName.String()
		} else {
			name += subName.String() + "-"
		}
	}
	sym := object.Symbol{
		Value: name,
	}
	lst := object.List{
		Value: []object.Object{&sym},
	}

	inner_code.Value = append(inner_code.Value, &lst)
	return code
}

func isDone(name_obj object.Object, code_obj object.Object) object.Object {
	code, ok := code_obj.(*object.List)
	if !ok {
		return newError("is_done expects second argument (code) to be a list, got %s", code_obj.Type())
	}

	names, ok := name_obj.(*object.List)
	if !ok {
		return newError("is_done expects first argument to be a list, got %s", code_obj.Type())
	}

	name := ""
	for i, n := range names.Value {
		if i == len(names.Value)-1 {
			name += n.String()
		} else {
			name += n.String() + "-"
		}
	}

	for i, block := range code.Value {
		// Skip over header
		if i == 0 {
			continue
		}

		l, ok := head(block).(*object.Symbol)
		if !ok {
			continue
		}
		if name == l.Value {
			return TRUE
		}
	}
	return FALSE
}

func CallPrimitive(name string, args []object.Object) object.Object {
	switch name {
	case "hd":
		if len(args) != 1 {
			return newError("hd takes one input, got %d", len(args))
		}
		return head(args[0])
	case "tl":
		if len(args) != 1 {
			return newError("tl takes one input, got %d", len(args))
		}
		return tail(args[0])
	case "o":
		if len(args) < 2 {
			return newError("o takes at least 2 inputs, got %d", len(args))
		}
		return o(args[0], args[1:]...)
	case "list":
		return list(args...)
	case "cons":
		if len(args) != 2 {
			return newError("cons takes two inputs, got %d", len(args))
		}
		return cons(args[0], args[1])
	case "newTail":
		if len(args) != 2 {
			return newError("newTail takes two inputs, got %d", len(args))
		}
		return newTail(args[0], args[1])
	case "newHeader":
		if len(args) < 1 {
			return newError("newHeader takes at least 1 input, got %d", len(args))
		}
		return newHeader(args[0], args[1:]...)
	case "newBlock":
		if len(args) != 2 {
			return newError("newBlock takes two inputs, got %d", len(args))
		}
		return newBlock(args[0], args[1])
	case "isDone":
		if len(args) != 2 {
			return newError("isDone takes two inputs, got %d", len(args))
		}
		return isDone(args[0], args[1])
	default:
		return newError("undefined primitive %s", name)
	}
}
