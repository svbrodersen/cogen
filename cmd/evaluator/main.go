package main

import (
	"cogen/evaluator"
	"cogen/lexer"
	"cogen/object"
	"cogen/parser"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

func printParserErrors(out io.Writer, errors []parser.ParserError) {
	for _, err := range errors {
		io.WriteString(out, "\t"+err.Msg+"\n")
	}
}

func fail(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "got error: %v", err)
	}
	fmt.Fprintf(os.Stderr, "usage: %s [inputfile] [delta]", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func parseCLIArgument(arg string) object.Object {
	if arg == "true" {
		return evaluator.TRUE
	}
	if arg == "false" {
		return evaluator.FALSE
	}
	if i, err := strconv.ParseInt(arg, 0, 64); err == nil {
		return &object.Integer{Value: i}
	}
	return &object.Symbol{Value: arg}
}

func main() {
	argc := len(os.Args)
	if argc < 2 {
		fail(nil)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fail(err)
	}

	prog := string(data)
	l := lexer.New(prog)
	p := parser.New(l)
	env := object.NewEnvironment()

	// Parse program and check for errors
	parsed_program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		io.WriteString(os.Stdout, p.GetErrorMessage())
		return
	}

	if len(parsed_program.Variables) > 0 {
		expectedArgs := len(parsed_program.Variables)
		if len(os.Args) < 2+expectedArgs {
			fmt.Fprintf(os.Stderr, "Program expects %d arguments, got %d\n", expectedArgs, len(os.Args)-2)
			os.Exit(1)
		}

		for i, input := range parsed_program.Variables {
			arg := os.Args[2+i]
			val := parseCLIArgument(arg)
			env.Set(input.Ident.Value, val)
		}
	}

	e := evaluator.New(parsed_program)
	evaluated := e.Eval(parsed_program, env)
	if evaluated != nil {
		io.WriteString(os.Stdout, fmt.Sprintf("Result: %s\n", evaluated.String()))
	} else {
		io.WriteString(os.Stdout, "evaluated is nil\n")
	}
}
