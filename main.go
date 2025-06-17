package main

import (
	"cogen/generator"
	"cogen/lexer"
	"cogen/parser"
	"flag"
	"fmt"
	"os"
	"strconv"
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

	delta := []int{}
	if argc != 2 {
		i := 2
		delta := make([]int, argc-2)
		for i < argc {
			val, err := strconv.ParseInt(os.Args[i], 10, 64)
			if err != nil {
				fail(err)
			}
			delta[i-2] = int(val)
			i += 1
		}
	}

	prog := string(data)
	l := lexer.New(prog)
	p := parser.New(l)
	c := generator.New(p)
	got := c.Gen(delta)
	fmt.Println(got)
}
