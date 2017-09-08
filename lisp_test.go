package main

import (
	"fmt"
	"testing"

	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/env"
	"github.com/steinarvk/heisenlisp/gen/parser"
	"github.com/steinarvk/heisenlisp/types"
)

func TestExpressionsTruthy(t *testing.T) {
	root := env.New(nil)
	builtin.BindDefaults(root)

	exprs := []string{
		"123",
		"(+ 1 234)",
		"#true",
		`"hello"`,
		"(if #true #true #false)",
		"(if 42 #true #false)",
		"(+ 1 0)",
		"(- 0 -47)",
		"(not 0)",
		"(not false)",
		"(not #false)",
		"(not (- 1 1))",
		"(not (+ 42 -42))",
		"(not (+ 42 (- 42)))",
		"(not (- 42 (set! my-special-symbol 42)))",
		"(not (- my-special-symbol 42))",
		"(= 42 (+ 1 41))",
		"(= 8 (* 2 2 2))",
		"(= 120 (* 2 3 4 5))",
		"(= 1307674368000 (* 2 3 4 5 6 7 8 9 10 11 12 13 14 15))",
		"(= -1307674368000 (* 2 3 4 5 6 7 8 9 10 -11 12 13 14 15))",
		`(defun! f (x y z) (- x (* y z)))
		 (= (f 2 3 4) -10)`,
		"(_atom-eq? ((lambda (z y x) (+ x (* y z))) 2 3 4) 10)",
		"(_atom? 42)",
		`(_atom? "hello")`,
		"(not (_atom? (lambda (x) x)))",
		"(_atom? (quote 42))",
		"(not (_atom? (quote (quote 42))))",
		`(_atom-eq? (+ 42 95) (let ((x 42) (y 95)) (+ x y)))`,
		`(_atom-eq? 1
		  (let ((x 2) (y 3))
		    (let ((x y) (y x))
			    (- x y))))`,
		"(_atom-eq? (_type 42) (quote integer))",
		"(_atom-eq? (_type \"kitten\") (quote string))",
		"(_atom-eq? (_type nil) (quote nil))",
		"(_atom-eq? (_type (cons 1 2)) (quote cons))",
		`(let ((f (lambda (&optional (x 3) (y x)) (+ x y))))
		   (_atom-eq? 6 (f)))`,
		`(let ((f (lambda (&optional (x 3) (y x)) (+ x y))))
		   (_atom-eq? 16 (f 8)))`,
		`(let ((f (lambda (&optional (x 3) (y x)) (+ x y))))
		   (_atom-eq? 17 (f 8 9)))`,
	}

	for i, s := range exprs {
		rv, err := parser.Parse(fmt.Sprintf("<testcase %d>", i), []byte(s))
		if err != nil {
			t.Errorf("error parsing #%d %q: %v", i, s, err)
			continue
		}

		exprs, ok := rv.([]interface{})
		if !ok {
			t.Errorf("error parsing #%d %q: %v", i, s, err)
		}

		var result types.Value
		for j, xpr := range exprs {
			result, err = xpr.(types.Value).Eval(root)
			if err != nil {
				t.Errorf("error evaluating #%d %q (sub-expression %d: %v): %v", i, s, j, xpr, err)
				break
			}
		}
		if err != nil {
			continue
		}

		if result.Uncertain() {
			t.Errorf("uncertain result for #%d %q: %v", i, s, result)
			continue
		}

		if result.Falsey() {
			t.Errorf("falsey result for #%d %q: %v", i, s, result)
		}
	}
}
