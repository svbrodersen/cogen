package evaluator

import (
	"bytes"
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
func o[T object.ValueString](s1 T, s2 T) object.Object {
	return &object.String{Value: s1.GetValue() + s2.GetValue() + "\n"}
}

// Should create a object.List
func list[T object.ValueString](a []T) object.Object {
	var out bytes.Buffer
	out.WriteString("(")
	for i, elem := range a {
		if i == len(a)-1 {
			out.WriteString(elem.GetValue())
		} else {
			out.WriteString(elem.GetValue() + " ")
		}
	}
	out.WriteString(")")
	return &object.String{Value: out.String()}
}

func CallFunction(name string, args []object.Object) object.Object {
	switch name {
	case "hd":
		if len(args) != 1 {
			return newError("function hd takes one input, got %d", len(args))
		}
		input, ok := args[0].(*object.List)
		if !ok {
			return newError("function hd expects list, got %s", args[0].Type())
		}
		return head(input)
	case "tl":
		if len(args) != 1 {
			return newError("function hd takes one input, got %d", len(args))
		}
		input, ok := args[0].(*object.List)
		if !ok {
			return newError("function hd expects list, got %s", args[0].Type())
		}
		return tail(input)
	case "o":
		if len(args) != 2 {
			return newError("function o takes 2 input, got %d", len(args))
		}
		item1, ok := args[0].(object.ValueString)
		if !ok {
			return newError("o got unexpected %s", args[0].Type())
		}

		item2, ok := args[1].(object.ValueString)
		if !ok {
			return newError("o got unexpected %s", args[1].Type())
		}
		return o(item1, item2)
	case "list":
		var items []object.ValueString
		for i, it := range args {
			item, ok := it.(object.ValueString)
			if !ok {
				return newError("list got unexpected %s", it.Type())
			}
			items[i] = item
		}
		return list(items)
	default:
		return newError("undefined function %s", name)
	}
}
