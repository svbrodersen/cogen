package main

import (
	"cogen/lexer"
	"cogen/parser"
	"flag"
	"fmt"
	"os"
)

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

	// Parse program and check for errors
	parsed_program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println(p.GetErrorMessage())
	} else {
		fmt.Println(parsed_program.String())
	}
}
