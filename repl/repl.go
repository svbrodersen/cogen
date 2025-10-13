package repl

import (
	"bufio"
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

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
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

		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}
		e := evaluator.New(program)
		evaluated := e.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}
