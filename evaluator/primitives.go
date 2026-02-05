package evaluator

import (
	"cogen/object"
	"fmt"
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

func isTerminator(val string) bool {
	return val == "if" || val == "goto" || val == "return"
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

// o appends inputs to the top-most block on the stack.
// If the appended instruction starts with if/goto/return, the block is popped and added to the Result Program.
func o(code_obj object.Object, inputs ...object.Object) object.Object {
	code, ok := code_obj.(*object.List)
	if !ok {
		return newError("o expected first argument to be list, got %s", code_obj.Type())
	}
	if len(code.Value) < 2 {
		// Expect at least [ResultList, ActiveBlock]
		// If len is 1, we are trying to append to the ResultList directly or stack is empty
		return newError("o called with empty stack (no active block to append to)")
	}

	// 1. Get the Active Block (Top of stack / Last element)
	stackTopIdx := len(code.Value) - 1
	activeBlock, ok := code.Value[stackTopIdx].(*object.List)
	if !ok {
		return newError("o expected top of stack to be a list, got %s", code.Value[stackTopIdx].Type())
	}

	// 2. Append inputs to the Active Block
	activeBlock.Value = append(activeBlock.Value, inputs...)

	// 3. Check if we need to "Finish" this block
	// We check the first input provided to see if it is a terminator instruction
	if len(inputs) > 0 {
		// The instruction is likely a List (e.g., ('return ...))
		if instr, ok := inputs[0].(*object.List); ok && len(instr.Value) > 0 {
			// Check the head of the instruction
			if sym, ok := head(instr).(*object.Symbol); ok {
				if isTerminator(sym.Value) {
					// 4. POP the active block from the stack
					code.Value = code.Value[:stackTopIdx]

					// 5. ADD it to the Result Program (Index 0)
					resProg, ok := code.Value[0].(*object.List)
					if !ok {
						return newError("o expected index 0 to be Result Program List")
					}
					resProg.Value = append(resProg.Value, activeBlock)
				}
			}
		}
	}

	return code
}

// newTail(2, '((0 if 0 goto 3) (1 right) (2 goto 0) (3 write 1)))
func newTail(item object.Object, Q_obj object.Object) object.Object {
	Q, ok := Q_obj.(*object.List)
	if !ok {
		return newError("newTail expects second element to be a list, got %s", Q_obj.Type())
	}
	val := item.String()
	i := 0
	for _, block := range Q.Value {
		lst, ok := block.(*object.List)
		if !ok {
			return newError("newTail expects second input to be list of list, got %s", block.Type())
		}
		if len(lst.Value) == 0 {
			continue
		}
		// We only search for symbol statements
		v, ok := lst.Value[0].(object.ValueString)
		if !ok {
			return newError("newTail expects the first value of each sublist to implement the ValueString interface.")
		}
		if v.GetValue() == val {
			break
		}
		i++
	}
	return &object.List{Value: Q.Value[i:]}
}

// newHeader initializes the code structure: [ [HeaderBlock] ]
func newHeader(name_obj object.Object, dynVars ...object.Object) object.Object {
	v, ok := name_obj.(*object.List)
	if !ok {
		return newError("newHeader expects first argument to be a list, got %s", name_obj.Type())
	}

	// 1. Create the Header Block content
	headerBlock := object.List{Value: []object.Object{}}

	var name string
	for i, subName := range v.Value {
		if i == len(v.Value)-1 {
			name += subName.String()
		} else {
			name += subName.String() + "_"
		}
	}
	sym := object.Symbol{Value: name}
	headerBlock.Value = append(headerBlock.Value, &sym)
	headerBlock.Value = append(headerBlock.Value, dynVars...)

	// 2. Create the Result Program List containing this Header Block
	resProg := object.List{
		Value: []object.Object{&headerBlock},
	}

	// 3. Return the container: [ resProg ]
	// Note: We do NOT have an active block on the stack yet.
	// The stack is represented by indices 1+. Currently, len is 1.
	container := object.List{
		Value: []object.Object{&resProg},
	}

	return &container
}

// newBlock pushes a new active block onto the stack
func newBlock(code_obj object.Object, name_obj object.Object) object.Object {
	code, ok := code_obj.(*object.List)
	if !ok {
		return newError("newBlock expects first argument (code) to be a list, got %s", code_obj.Type())
	}

	name_list, ok := name_obj.(*object.List)
	if !ok {
		return newError("newBlock expects second argument to be a list, got %s", name_list.Type())
	}

	// 1. Construct the Name
	name := ""
	for i, subName := range name_list.Value {
		if i == len(name_list.Value)-1 {
			name += subName.String()
		} else {
			name += subName.String() + "_"
		}
	}

	// 2. Create the new Active Block (starting with the label)
	sym := object.Symbol{Value: name}
	activeBlock := object.List{
		Value: []object.Object{&sym},
	}

	// 3. Push to Stack (Append to code)
	code.Value = append(code.Value, &activeBlock)

	return code
}

// isDone checks if a block exists in the Result Program (code[0])
func isDone(name_obj object.Object, code_obj object.Object) object.Object {
	code, ok := code_obj.(*object.List)
	if !ok {
		return newError("is_done expects second argument (code) to be a list, got %s", code_obj.Type())
	}

	// Parse the target name
	names, ok := name_obj.(*object.List)
	if !ok {
		return newError("is_done expects first argument to be a list, got %s", name_obj.Type())
	}
	name := ""
	for i, n := range names.Value {
		if i == len(names.Value)-1 {
			name += n.String()
		} else {
			name += n.String() + "_"
		}
	}

	// Helper to check a specific block for the label
	checkBlock := func(blockObj object.Object) bool {
		// Blocks are Lists
		block, ok := blockObj.(*object.List)
		if !ok || len(block.Value) == 0 {
			return false
		}
		// The first element of a block is its Label (Symbol)
		labelSym, ok := block.Value[0].(*object.Symbol)
		if !ok {
			return false
		}
		return labelSym.Value == name
	}

	// 1. Check the Result Program (code[0])
	// This is a List of Lists (Finished Blocks)
	if len(code.Value) > 0 {
		if resProg, ok := code.Value[0].(*object.List); ok {
			for _, block := range resProg.Value {
				if checkBlock(block) {
					return TRUE
				}
			}
		}
	}

	// 2. Check the Active Stack (code[1:])
	// These are individual Lists (Active Blocks)
	if len(code.Value) > 1 {
		for _, block := range code.Value[1:] {
			if checkBlock(block) {
				return TRUE
			}
		}
	}

	return FALSE
}

func cleanOutput(code_obj object.Object) object.Object {
	code, ok := code_obj.(*object.List)
	if !ok {
		return newError("is_done expects second argument (code) to be a list, got %s", code_obj.Type())
	}

	prog, err := ConvertSExprToAST(code.Value)
	if err != nil {
		return newError("cleanOutput failed. Got input: %s\n\n Failed with error %s", code_obj.String(), err)
	}

	fmt.Println(prog.String())
	return &object.Integer{Value: 0}
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
	case "cleanOutput":
		if len(args) != 1 {
			return newError("cleanOutput expects a single output, got %d", len(args))
		}
		return cleanOutput(args[0])
	default:
		return newError("undefined primitive %s", name)
	}
}
