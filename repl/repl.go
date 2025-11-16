package repl

import (
	"bufio"
	"bytes"
	"cogen/ast"
	"cogen/evaluator"
	"cogen/lexer"
	"cogen/object"
	"cogen/parser"
	"fmt"
	"io"
	"regexp"
)

const PROMPT = ">> "

func printParserErrors(out io.Writer, errors []parser.ParserError) {
	for _, err := range errors {
		io.WriteString(out, "\t"+err.Msg+"\n")
	}
}
func isDuplicate(a, b *ast.LabelStatement) bool {
	return a.Label == b.Label
}

func unionProgram(orig, new *ast.Program) *ast.Program {
	result := make([]*ast.LabelStatement, len(orig.Statements))
	copy(result, orig.Statements)

	for _, newItem := range new.Statements {
		replaced := false
		for i, origItem := range orig.Statements {
			if isDuplicate(origItem, newItem) {
				result[i] = newItem
				replaced = true
				break
			}
		}
		if !replaced {
			result = append(result, newItem)
		}
	}
	orig.Statements = result
	return orig
}

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	var input_string bytes.Buffer
	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		// Look for a label start, but not assign :=
		pattern := `^[^:=]+:[^=].*$`
		matched, err := regexp.MatchString(pattern, line)
		if err != nil {
			panic(err)
		}
		// If not found, then add the starting label to be happy.
		if !matched {
			line = "1: " + line
		}
		input_string.WriteString(line)

		l := lexer.New(input_string.String())
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}
		e := evaluator.New(program)
		evaluated := e.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.String())
			io.WriteString(out, "\n")
		}
	}
}
