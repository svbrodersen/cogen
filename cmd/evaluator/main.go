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
		printParserErrors(os.Stdout, p.Errors())
		return
	}
	e := evaluator.New(parsed_program)
	evaluated := e.Eval(parsed_program, env)
	if evaluated != nil {
		io.WriteString(os.Stdout, fmt.Sprintf("Result: %s\n", evaluated.Inspect()))
	} else {
		io.WriteString(os.Stdout, "evaluated is nil\n")
	}
}
