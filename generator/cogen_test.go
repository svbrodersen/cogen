package generator_test

import (
	"cogen/generator"
	"cogen/lexer"
	"cogen/parser"
	"log"
	"strings"
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
			name: "Ackermann",
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
			want: `ack(m);
0-0-m: code := newheader(n, (list '0 m));
        goto 1-0-m;
2: return code;
1-ack-m: if done?((list 'ack m), code) 2 3-ack-m;
3-ack-m: code := newblock(code, (list 'ack m));
        goto 4-ack-m;
4-ack-m: if (m = 0) 4-done-m 4-next-m;
4-done-m: return o(code, (list 'return (list ('n + '1))));
4-next-m: code := call 1-ack0-m;
        code := call 1-ack1-m;
        return o(code, (list '(list ('n = '0)) (list 'ack0 m) (list 'ack1 m)));
1-ack0-m: if done?((list 'ack0 m), code) 2 3-ack0-m;
3-ack0-m: code := newblock(code, (list 'ack0 m));
        goto 4-ack0-m;
4-ack0-m: n := 1;
        m := (m - 1);
        n := call 5-ack-m-n;
        return o(code, (list 'return (list 'quote n)));
5-ack-m-n: if (m = 0) 5-done-m-n 5-next-m-n;
5-done-m-n: return (n + 1);
5-next-m-n: if (n = 0) 5-ack0-m-n 5-ack1-m-n;
5-ack0-m-n: n := 1;
        goto 5-ack2-m-n;
5-ack2-m-n: m := (m - 1);
        n := call ack m n;
        return n;
5-ack1-m-n: n := (n - 1);
        n := call ack m n;
        goto 5-ack2-m-n;
1-ack1-m: if done?((list 'ack1 m), code) 2 3-ack1-m;
3-ack1-m: code := newblock(code, (list 'ack1 m));
        goto 4-ack1-m;
4-ack1-m: code := o(code, (list 'n ':= (list ('n - '1))));
        code := call 1-ack-m;
        code := o(code, (list 'n ':= (list 'call (list 'ack m))));
        m := (m - 1);
        code := call 1-ack-m;
        code := o(code, (list 'n ':= (list 'call (list 'ack m))));
        return o(code, (list 'return 'n));`,
		}, {
			name: "Assign",
			prog: `
init (s, ta, tb);
init: q := '0;
		goto loop;
loop: if (s = '()) end else isab;
isab: c := car(s);
		s := cdr(s);
		if (c = 'a) doa else dob;
doa: q := (ith(ta, q));
		goto loop;
dob: q := (ith(tb,q));
		goto loop;
end: return q;
`,
			want: `init(s);
0-0-s: code := newheader(ta, tb, (list '0 s));
	goto 1-0-s;
2: return code;
1-init-s: if done?((list 'init s), code) 2 3-init-s;
3-init-s: code := newblock(code, (list 'init s));
	goto 4-init-s;
4-init-s: q := '0;
if (s = '()) 4-end-q-s 4-isab-q-s;
4-end-q-s: return o(code, (list 'return (list 'quote q)));
4-isab-q-s: c := car(s);
	s := cdr(s);
	code := call 1-doa-c-q-s;
	code := call 1-dob-c-q-s;
	return o(code, (list '(list ((list 'quote c) = ''a)) (list 'doa q c s) (list 'dob s q c)));
1-doa-c-q-s: if done?((list 'doa s q c), code) 2 3-doa-c-q-s;
3-doa-c-q-s: code := newblock(code, (list 'doa s q c));
	goto 4-doa-c-q-s;
4-doa-c-q-s: code := o(code, (list 'q ':= ith()));
if (s = '()) 4-end-c-s 4-isab-c-s;
4-end-c-s: return o(code, (list 'return 'q));
4-isab-c-s: c := car(s);
s := cdr(s);
code := call 1-doa-c-s;
code := call 1-dob-c-s;
return o(code, (list '(list ((list 'quote c) = ''a)) (list 'doa s c) (list 'dob s c)));
1-doa-c-s: if done?((list 'doa s c), code) 2 3-doa-c-s;
3-doa-c-s: code := newblock(code, (list 'doa s c));
goto 4-doa-c-s;
4-doa-c-s: code := o(code, (list 'q ':= ith()));
if (s = '()) 4-end-c-s 4-isab-c-s;
1-dob-c-s: if done?((list 'dob s c), code) 2 3-dob-c-s;
3-dob-c-s: code := newblock(code, (list 'dob s c));
goto 4-dob-c-s;
4-dob-c-s: code := o(code, (list 'q ':= ith()));
if (s = '()) 4-end-c-s 4-isab-c-s;
1-dob-c-q-s: if done?((list 'dob s q c), code) 2 3-dob-c-q-s;
3-dob-c-q-s: code := newblock(code, (list 'dob s q c));
goto 4-dob-c-q-s;
4-dob-c-q-s: code := o(code, (list 'q ':= ith()));
if (s = '()) 4-end-c-s 4-isab-c-s;`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.prog)
			p := parser.New(l)
			c := generator.New(p)
			got, _ := c.Gen([]int{0})
			log.Printf("Program:\n%s", c.OriginalProgram)
			log.Printf("Cogen:\n%s", got)
			if strings.EqualFold(got.String(), tt.want) {
				t.Errorf("Expected: \n%s\n Got: \n%s\n", tt.want, got)
			}
		})
	}
}
