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
	return &object.List{Value: s.Value[1:(len(s.Value) - 1)]}
}

// code should be a Object list, and we add to the first element.
// TODO: Causes panic atm
func o(s1 *object.List, inputs ...object.Object) object.Object {
	val := append(s1.Value, inputs...)
	return &object.List{Value: val}
}

// Should create a object.List
func list(a ...object.Object) object.Object {
	return &object.List{Value: a}
}

func new_tail(item object.Object, Q *object.List) object.Object {
	val := item.Inspect()
	val += ":"
	i := 0
	for _, block := range Q.Value {
		lst, ok := block.(*object.List)
		if !ok {
			return newError("new_tail expects second input to be list of list, got %s", block.Type())
		}
		if lst.Value[0].Inspect() == val {
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
			return newError("hd takes one input, got %d", len(args))
		}
		input, ok := args[0].(*object.List)
		if !ok {
			return newError("hd expects list, got %s", args[0].Type())
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
		if len(args) < 1 {
			return newError("list expected atleast one input, got %d", len(args))
		}
		return list(args...)
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
