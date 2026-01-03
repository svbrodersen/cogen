package evaluator

import (
	"cogen/object"
)

func head(s *object.List) object.Object {
	if len(s.Value) == 0 {
		return newError("hd called on empty list")
	}
	return s.Value[0]
}

// Should be all but first element
func tail(s *object.List) object.Object {
	if len(s.Value) == 0 {
		return newError("tl called on empty list")
	}
	return &object.List{Value: s.Value[1:]}
}

func o(s1 *object.List, inputs ...object.Object) object.Object {
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

// new_tail(2, '((0 if 0 goto 3) (1 right) (2 goto 0) (3 write 1)))
func new_tail(item object.Object, Q *object.List) object.Object {
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

func CallPrimitive(name string, args []object.Object) object.Object {
	switch name {
	case "hd":
		if len(args) != 1 {
			return newError("hd takes one input, got %d", len(args))
		}
		input, ok := args[0].(*object.List)
		if !ok {
			return newError("hd expects list, got %s", args[0].Type())
		}
		return head(input)
	case "tl":
		if len(args) != 1 {
			return newError("tl takes one input, got %d", len(args))
		}
		input, ok := args[0].(*object.List)
		if !ok {
			return newError("tl expects list, got %s", args[0].Type())
		}
		return tail(input)
	case "o":
		if len(args) < 2 {
			return newError("o takes at least 2 inputs, got %d", len(args))
		}
		item1, ok := args[0].(*object.List)
		if !ok {
			return newError("o expected first argument to be list, got %s", args[0].Type())
		}
		return o(item1, args[1:]...)
	case "list":
		return list(args...)
	case "cons":
		if len(args) != 2 {
			return newError("cons takes two inputs, got %d", len(args))
		}
		return cons(args[0], args[1])
	case "new_tail":
		if len(args) != 2 {
			return newError("new_tail takes two inputs, got %d", len(args))
		}
		input, ok := args[1].(*object.List)
		if !ok {
			return newError("new_tail expects second element to be a list, got %s", args[1].Type())
		}
		return new_tail(args[0], input)
	default:
		return newError("undefined primitive %s", name)
	}
}
