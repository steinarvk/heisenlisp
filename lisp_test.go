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
	}

	for i, s := range exprs {
		rv, err := parser.Parse(fmt.Sprintf("<testcase %d>", i), []byte(s))
		if err != nil {
			t.Errorf("error parsing #%d %q: %v", i, s, err)
			continue
		}

		evaled, err := rv.(types.Value).Eval(root)
		if err != nil {
			t.Errorf("error evaluating #%d %q: %v", i, s, err)
			continue
		}

		if evaled.Uncertain() {
			t.Errorf("uncertain result for #%d %q: %v", i, s, evaled)
			continue
		}

		if evaled.Falsey() {
			t.Errorf("falsey result for #%d %q: %v", i, s, evaled)
		}
	}
}
