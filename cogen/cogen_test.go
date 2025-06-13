package cogen_test

import (
	"cogen/cogen"
	"cogen/lexer"
	"cogen/parser"
	"log"
	"testing"
)

func TestCogen_Gen(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		prog string
		want string
	}{
		{
			name: "Assign",
			prog: `
			ack (m, n);
			ack: if m = 0 goto done else next;
			next: if n = 0 goto ack0 else ack1;
			done: return n + 1;
			ack0: n := 1;
						goto ack2;
			ack1: n := n - 1;
						n := call ack m n;
						goto ack2;
			ack2: m := m - 1;
						n := call ack m n;
						return n;`,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.prog)
			p := parser.New(l)
			c := cogen.New(p)
			got := c.Gen([]int{0})
			log.Println(got)
		})
	}
}
