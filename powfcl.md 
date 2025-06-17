## Original

pow(m, n);
init: result := 1;
      goto test;
test: if n < 1 goto end else loop;
loop: result := result * m;
      n := n - 1;
      goto test;
end: return result;

## Codegen for m static
pow(m);
0-init-m: code := newheader(n, list('pow, m));
        goto 1-init-m;
2: return code;
1-init-m: if done?(list('init, m), code) 2 3-init-m;
3-init-m: code := newblock(code, list('init, m));
        goto 4-init-m;
4-init-m: result := 1;
        code := call 1-end-m-result;
        code := call 1-loop-m-result;
        return o(code, list('if, list('n, '<, '1), list('end, m, result), list('loop, m, result)));
1-end-m-result: if done?(list('end, m, result), code) 2 3-end-m-result;
3-end-m-result: code := newblock(code, list('end, m, result));
        goto 4-end-m-result;
4-end-m-result: return o(code, list('return, list('quote, result)));
1-loop-m-result: if done?(list('loop, m, result), code) 2 3-loop-m-result;
3-loop-m-result: code := newblock(code, list('loop, m, result));
        goto 4-loop-m-result;
4-loop-m-result: result := (result * m);
        code := o(code, list('n, ':=, list('n, '-, '1)));
        code := call 1-end-m-result;
        code := call 1-loop-m-result;
        return o(code, list('if, list('n, '<, '1), list('end, m, result), list('loop, m, result)));

## code with m = 2
pow-2(n);
init-2:  // never finishes
end-2-1: return 1;
end-2-2: return 2;
end-2-4: return 4;
loop-2-4: n := n - 1;
          if n < 1 end-2-8 else loop-2-8;
loop-2-2: n := n - 1;
          if n < 1 end-2-4 else loop-2-4;
loop-2-1: n := n - 1;
          if n < 1 end-2-2 else loop-2-2;

## Codegen for n static
pow(n);
0-init-n: code := newheader(m, list('pow, n));
        goto 1-init-n;
2: return code;
1-init-n: if done?(list('init, n), code) 2 3-init-n;
3-init-n: code := newblock(code, list('init, n));
        goto 4-init-n;
4-init-n: result := 1;
        if (n < 1) 4-end-n-result 4-loop-n-result;
4-end-n-result: return o(code, list('return, list('quote, result)));
4-loop-n-result: code := o(code, list('result, ':=, list(list('quote, result), '*, 'm)));
        n := (n - 1);
        if (n < 1) 4-end-n 4-loop-n;
4-end-n: return o(code, list('return, 'result));
4-loop-n: code := o(code, list('result, ':=, list('result, '*, 'm)));
        n := (n - 1);
        if (n < 1) 4-end-n 4-loop-n;

## code with n=2
pow-2(m);
init-2: result := 1 * m
        result := result * m
        return result


## Codegen both Static
pow(m, n);
0-init-m-n: code := newheader(list('pow, m, n));
        goto 1-init-m-n;
2: return code;
1-init-m-n: if done?(list('init, m, n), code) 2 3-init-m-n;
3-init-m-n: code := newblock(code, list('init, m, n));
        goto 4-init-m-n;
4-init-m-n: result := 1;
        if (n < 1) 4-end-m-n-result 4-loop-m-n-result;
4-end-m-n-result: return o(code, list('return, list('quote, result)));
4-loop-m-n-result: result := (result * m);
        n := (n - 1);
        if (n < 1) 4-end-m-n-result 4-loop-m-n-result;

## n=2, m=2
pow-2-2():
init-2-2: return '4

## Codegen both dynamic
pow();
0-init: code := newheader(m, n, list('pow));
        goto 1-init;
2: return code;
1-init: if done?(list('init), code) 2 3-init;
3-init: code := newblock(code, list('init));
        goto 4-init;
4-init: result := 1;
        code := call 1-end-result;
        code := call 1-loop-result;
        return o(code, list('if, list('n, '<, '1), list('end, result), list('loop, result)));
1-end-result: if done?(list('end, result), code) 2 3-end-result;
3-end-result: code := newblock(code, list('end, result));
        goto 4-end-result;
4-end-result: return o(code, list('return, list('quote, result)));
1-loop-result: if done?(list('loop, result), code) 2 3-loop-result;
3-loop-result: code := newblock(code, list('loop, result));
        goto 4-loop-result;
4-loop-result: code := o(code, list('result, ':=, list(list('quote, result), '*, 'm)));
        code := o(code, list('n, ':=, list('n, '-, '1)));
        code := call 1-end;
        code := call 1-loop;
        return o(code, list('if, list('n, '<, '1), list('end), list('loop)));
1-end: if done?(list('end), code) 2 3-end;
3-end: code := newblock(code, list('end));
        goto 4-end;
4-end: return o(code, list('return, 'result));
1-loop: if done?(list('loop), code) 2 3-loop;
3-loop: code := newblock(code, list('loop));
        goto 4-loop;
4-loop: code := o(code, list('result, ':=, list('result, '*, 'm)));
        code := o(code, list('n, ':=, list('n, '-, '1)));
        code := call 1-end;
        code := call 1-loop;
        return o(code, list('if, list('n, '<, '1), list('end), list('loop)));

# Would create:
pow(m, n):
init: if n < 1 end-1 loop-1
end-1: return 1
loop-1: result := 1 * m
        if n < 1 end loop
end: return result
loop: result := result * m
      n := n - 1
      if n < 1 end loop
