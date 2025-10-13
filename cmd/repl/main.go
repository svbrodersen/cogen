package main

import (
	"cogen/repl"
	"fmt"
	"os"
)

func main() {
	fmt.Printf("Feel free to type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
}
