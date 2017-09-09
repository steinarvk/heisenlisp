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
		`(defmacro! and-two (a b)
		   (let ((s (cons (quote if) (cons b (cons true (cons false nil))))))
			   (cons (quote if) (cons a (cons s (cons false nil))))))`,
		"(_atom-eq? true (and-two true true))",
		"(_atom-eq? false (and-two false true))",
		"(_atom-eq? false (and-two true false))",
		"(_atom-eq? false (and-two false false))",
		"(_atom-eq? true (and-two true (and-two true true)))",
		"(_atom-eq? false (and-two true (and-two false true)))",
		"(_atom-eq? false (and-two true (and-two true false)))",
		"(_atom-eq? false (and-two false (and-two true true)))",
		"(defun! list (&rest xs) xs)",
		"(defun! cons? (x) (_atom-eq? (_type x) (quote cons)))",
		"(defun! nil? (x) (_atom-eq? (_type x) (quote nil)))",
		"(defun! second (x) (car (cdr x)))",
		`(defmacro! my-cond (clause &rest more-clauses)
		   (list 'if (car clause)
						  	 (second clause)
							 	 (if (nil? more-clauses)
										 nil
										 (cons 'my-cond more-clauses))))`,
		`(_atom-eq? 42
		   (my-cond (true 42)
			          (true 43)
								(true 44)))`,
		`(_atom-eq? 43
		   (my-cond ((_atom-eq? 2 3) 42)
			          (true 43)
								(true 44)))`,
		`(_atom-eq? 42
		   (my-cond ((_atom-eq? 2 2) 42)
			          (true 43)
								(true 44)))`,
		`(defun! simply-equal? (a b)
		  (my-cond
			  ((and-two (_atom? a) (_atom? b))
				 (_atom-eq? a b))
			  ((and-two (cons? a) (cons? b))
				 (and-two (simply-equal? (car a) (car b))
				 				  (simply-equal? (cdr a) (cdr b))))))`,
		`(simply-equal? (cons 1 2) (cons 1 2))`,
		`(not (simply-equal? (cons 1 2) (cons 1 3)))`,
		`(simply-equal? (cons 1 (cons 2 3)) (cons 1 (cons 2 3)))`,
		`(not (simply-equal? (cons 1 (cons 4 3)) (cons 1 (cons 2 3))))`,
		"(simply-equal? '(foo bar baz) '(FOO BAR BAZ))",
		`(let ((mjau '(list 1 2 3)))
       (simply-equal? (quasiquote (list 5 6 ,mjau))
			                '(list 5 6 (list 1 2 3))))`,
		`(let ((mjau '(list 1 2 3)))
       (simply-equal? (quasiquote (list 5 6 ,@mjau))
			                '(list 5 6 list 1 2 3)))`,
		`(let ((mjau '(list 1 2 3)))
       (simply-equal? ` + "`" + `(list 5 6 ,@mjau)
			                '(list 5 6 list 1 2 3)))`,
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
