package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/gen/parser"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/unknown"
)

func TestExpressionsTruthy(t *testing.T) {
	root := builtin.NewRootEnv()

	exprs := []string{
		"123",
		"(+ 1 234)",
		"true",
		`"hello"`,
		"(if true true false)",
		"(if 42 true false)",
		"(+ 1 0)",
		"(- 0 -47)",
		"(not 0)",
		"(not false)",
		"(not false)",
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
		`(simply-equal? (list 1 2 3) (possible-values (any-of 1 2 3)))`,
		`(nil? nil)`,
		`(not (if true nil 42))`,
		`(not (if false 42 nil))`,
		`(not (my-cond (false 95)))`,
		`(not (my-cond (false 38) (false 95)))`,
		`(not (simply-equal? (cons 1 nil) nil))`,
		`(simply-equal? (possible-values (any-of 1 2 3))
			              (possible-values (any-of 1 2 3)))`,
		`(not (simply-equal? (possible-values (any-of 1 2 3))
			                   (possible-values (any-of 1 2 3 4))))`,
		`(simply-equal? (possible-values (any-of (any-of 1 2) 3))
			              (possible-values (any-of 1 (any-of 2 3))))`,
		`(_atom-eq? 2 (if 1 2 3))`,
		`(_atom-eq? 3 (if 0 2 3))`,
		`(simply-equal? (list 42 43)
		                (possible-values (if (any-of true false) 42 43)))`,
		`(not (_atom-eq? true (any-of true false)))`,
		`(not (_atom-eq? false (any-of true false)))`,
		`(_atom-eq? true (not (_atom-eq? true (any-of true false))))`,
		`(simply-equal? (list 1 2) (possible-values (if (any-of true false) 1 2)))`,
		`(_atom-eq? "true" (_to-string true))`,
		`(_atom-eq? "false" (_to-string false))`,
		`(_atom-eq? "maybe" (_to-string (any-of true false)))`,
		`(_atom-eq? "maybe" (_to-string (any-of false true)))`,
		`(_atom-eq? "false" (_to-string false))`,
		`(_atom-eq? "(0 1 2 3 4 5)" (_to-string (range 6)))`,
		`(_atom-eq? "(1 2 3 4 5 6)" (_to-string (map inc (range 6))))`,
		`(_atom-eq? "(-1 0 1 2 3 4)" (_to-string (map dec (range 6))))`,
		`(_atom-eq? "1" (_to-string (reduce-left * 1 nil)))`,
		`(_atom-eq? "120" (_to-string (reduce-left * 1 (map inc (range 5)))))`,
		`(equals? (list 1 2 3) (list 1 2 3))`,
		`(not (equals? (list 1 2 3) (list 1 2 4)))`,
		`(may? maybe)`,
		`(not (may? false))`,
		`(may? true)`,
		`(not (must? maybe))`,
		`(not (must? false))`,
		`(must? true)`,
		`(must? 42)`,
		`(not (must? (list)))`,
		`(and true 43 95)`,
		`(not (and true 0 95))`,
		`(or true 0 95)`,
		`(not (or))`,
		`(and)`,
		`(and true true true)`,
		`(or true false true)`,
		`(not (and true false))`,
		`(not (or false false))`,
		`(may? (and true maybe))`,
		`(not (may? (and false maybe)))`,
		`(may? (and maybe maybe))`,
		`(may? (or maybe maybe))`,
		`(must? (or maybe true))`,
		`(not (must? (or maybe maybe)))`,
		`(any? (lambda (x) (= x 2)) (list 1 2 3))`,
		`(not (all? (lambda (x) (= x 2)) (list 1 2 3)))`,
		`(all? (lambda (x) (= x 2)) (list 2))`,
		`(all? (lambda (x) (and (= x 3) (= x 2))) (list))`,
		`(not (any? (lambda (x) (and (= x 3) (= x 2))) (list)))`,
		`(= (/ 4 2) (/ 8 4))`,
		`(= (when true 42 43) 43)`,
		`(= (when false 42 43) nil)`,
		`(= (unless true 42 43) nil)`,
		`(= (unless false 42 43) 43)`,
		`(<= 10 20)`,
		`(<= 10 10)`,
		`(>= 20 10)`,
		`(>= 20 20)`,
		`(< 10 20)`,
		`(not (< 10 10))`,
		`(> 20 10)`,
		`(not (> 20 20))`,
		`(not (<= 20 10))`,
		`(not (>= 10 20))`,
		`(_maybe? maybe)`,
		`(not (_maybe? true))`,
		`(not (_maybe? false))`,
		`(not (_maybe? "maybe"))`,
		`(not (_maybe? 'maybe))`,
		`(_maybe? (any-of true false))`,
		`(_maybe? (any-of false true))`,
		`(_maybe? (= (any-of 1 2) 2))`,
		`(_maybe? (= (unknown-of-type 'string) (unknown-of-type 'string)))`,
		`(not (= (unknown-of-type 'integer) (unknown-of-type 'string)))`,
		`(_maybe? (= (unknown-of-type 'string 'integer) (unknown-of-type 'string)))`,
		`(= 'string (type (unknown-of-type 'string)))`,
		`(= 'unknown-of-type (_type (type unknown)))`,
		`(= 'integer (type (any-of 1 2 3)))`,
		`(= 'symbol (type (type unknown)))`,
		`(= (list 'integer 'string) (possible-values (type (any-of 1 2 3 "hello"))))`,
		`(bool? true)`,
		`(bool? false)`,
		`(bool? maybe)`,
		`(function? (lambda (x) (+ 1 x)))`,
		`(function? type)`,
		`(function? string?)`,
		`(cons? (cons 1 2))`,
		`(integer? 42)`,
		`(macro? cond)`,
		`(nil? nil)`,
		`(floating-point? 3.14)`,
		`(string? "hello")`,
		`(symbol? 'hello)`,
		`(_maybe? (= (number-in-range 'from 100 'to 200) 200))`,
		`(not (= (number-in-range 'from 100 'below 200) 200))`,
		`(_maybe? (= (number-in-range 'from 100 'below 200) 150))`,
		`(_maybe? (= (number-in-range 'from 100 'below 200) 100))`,
		`(_maybe? (= (number-in-range 'from 100 'below 200) 100))`,
		`(not (= (number-in-range 'above 100 'below 200) 100))`,
		`(not (= (number-in-range 'above 100 'below 200) 50))`,
		`(_maybe? (= (number-in-range 'from 0 'to 1) (number-in-range 'from 0.5 'to 1.5)))`,
		`(_atom-eq? "#number-in-range([0,200])" (_to-string
		   (+ (number-in-range 'from 0 'to 100)
			    (number-in-range 'from 0 'to 100))))`,
		`(_atom-eq? "#number-in-range([100,200])" (_to-string
		   (+ (number-in-range 'from 50 'to 100)
			    (number-in-range 'from 50 'to 100))))`,
		`(_atom-eq? "#number-in-range((100,200])" (_to-string
		   (+ (number-in-range 'above 50 'to 100)
			    (number-in-range 'from 50 'to 100))))`,
		`(_atom-eq? "#number-in-range([100,200))" (_to-string
		   (+ (number-in-range 'from 50 'to 100)
			    (number-in-range 'from 50 'below 100))))`,
		`(_atom-eq? "#number-in-range([-51,58])" (_to-string
		   (- (number-in-range 'from 50 'to 100)
			    (number-in-range 'from 42 'to 101))))`,
		`(_atom-eq? "#number-in-range([2100,10100])" (_to-string
		   (* (number-in-range 'from 50 'to 100)
			    (number-in-range 'from 42 'to 101))))`,
		`(_atom-eq? "#number-in-range([2100,10100))" (_to-string
		   (* (number-in-range 'from 50 'below 100)
			    (number-in-range 'from 42 'to 101))))`,
		`(_atom-eq? "#number-in-range([-5050,10100])" (_to-string
		   (* (number-in-range 'from -50 'to 100)
			    (number-in-range 'from 42 'to 101))))`,
		`(_atom-eq? "#number-in-range([-13869,10050])" (_to-string
		   (* (number-in-range 'from -50 'to 69)
			    (number-in-range 'from -201 'to 101))))`,
		`(= 1 1.0)`,
		`(= 42 42.0)`,
		`(= (/ 1 2) 0.5)`,
		`(= "((((nil 1) 2) 3) 4)" (_to-string (fold-left list '() '(1 2 3 4))))`,
		`(= "(((1 2) 3) 4)" (_to-string (reduce-left list '() '(1 2 3 4))))`,
		`(= (length (list 1 2 3 4 5)) 5)`,
		`(> (number-in-range 'from 10 'to 20) 0)`,
		`(_maybe? (> (number-in-range 'from 11 'to 21) 16))`,
		`(not (> (number-in-range 'from 10 'to 20) 25))`,
		`(< 0 (number-in-range 'from 10 'to 20))`,
		`(_maybe? (< 15 (number-in-range 'from 10 'to 20)))`,
		`(not (< 25 (number-in-range 'from 10 'to 20)))`,
		`(_maybe? (<= (number-in-range 'from 10 'to 20) 10))`,
		`(not (<= (number-in-range 'above 10 'to 20) 10))`,
		`(_maybe? (>= (number-in-range 'from 12 'to 22) 22))`,
		`(not (>= (number-in-range 'from 10 'below 20) 20))`,
		`(_maybe? (= (number-in-range 'from 10 'to 20) 20))`,
		`(not (= (number-in-range 'from 10 'below 20) 20))`,
		`(_maybe? (= (number-in-range 'above 0) 5))`,
		`(not (= (number-in-range 'above 0) 0))`,
		`(_maybe? (= (number-in-range 'below 0) -5))`,
		`(not (= (number-in-range 'below 0) 5))`,
		`(_maybe? (= (any-of 0 1) 0))`,
		`(_maybe? (= (dec (any-of 0 1)) 0))`,
		`(not (= (dec (dec (any-of 0 1))) 0))`,
		`(not (= NaN NaN))`,
		`(odd? 3)`,
		`(even? 4)`,
		`(_maybe? (odd? (any-of 3 4)))`,
		`(_maybe? (even? (any-of 3 4)))`,
		`(not (odd? (any-of 6 4)))`,
		`(not (even? (any-of 3 5)))`,
		`(odd? (any-of 7 5))`,
		`(even? (any-of 8 10))`,
		`(_maybe? (contains? 6 (filter-reversed (lambda (n) (> n (any-of 5 6))) (range 10))))`,
		`(contains? 7 (filter-reversed (lambda (n) (> n (any-of 5 6))) (range 10)))`,
		`(not (contains? 5 (filter-reversed (lambda (n) (> n (any-of 5 6))) (range 10))))`,
		`(equals? (list 1 2 3) (reversed (list 3 2 1)))`,
		`(equals? nil (reversed nil))`,
		`(_maybe? (contains? 6 (filter (lambda (n) (> n (any-of 5 6))) (range 10))))`,
		`(contains? 7 (filter (lambda (n) (> n (any-of 5 6))) (range 10)))`,
		`(not (contains? 5 (filter (lambda (n) (> n (any-of 5 6))) (range 10))))`,
		`(equals? (list 1 2 3) (filter (is-not-equal-to 5) (list 1 5 2 3)))`,
		`(_maybe? (equals? (list 1 2 3) (filter (is-not-equal-to (any-of 5 6)) (list 1 5 2 3))))`,
		`(_maybe? (contains? 3 (map inc (filter (is-not-equal-to (any-of 2 10)) (list 1 2 3 4)))))`,
		`(= 0.5 (/ 2))`,
		`(= (factorial 5) 120)`,
		`(= (factorial 50) 30414093201713378043612608166064768844377641568960512000000000000)`,
		`(= 'integer (type (* (+ 1 (* (/ 400) 40090)) 40)))`,
		`(= 'integer (type (* 3/14 14)))`,
		`(= 'floating-point (type (to-floating-point 3/14)))`,
		`(not (= 'floating-point (type 3/14)))`,
		`(= (factorial 50) (* (/ (factorial 50) 123456789101232) 123456789101232))`,
		`(= 42 (handle-exception 42 ((x) (list x x))))`,
		`(= (list "foo" "foo") (handle-exception (throw-exception "foo") ((x) (list x x))))`,
		`(= (list (list 1 2) (list 1 3)) (possible-values (any-of (list 1 2) (list 1 2) (list 1 3))))`,
		`(= (list 1 2 3 4 5 6) (list* 1 2 3 (list 4 5 6)))`,
		`(= (letfunc ((f (x y z) (+ y (* x z)))) (f 2 3 4)) 11)`,
		`(= (list 1 2 3 4 5 6) (append (list 1 2) (list 3 4) (list 5 6)))`,
		`(= 3 (let* ((x 0) (y (inc x)) (z (inc y)) (z (+ z y))) z))`,
		`(= (list 1 2 3 4 5 6 7 8 9) (sorted (list 7 3 2 1 5 4 9 8 6)))`,
		`(= (list 1 2 3 4 5) (sorted (list 4 3 2 5 1)))`,
		`(= (list) (sorted nil))`,
		`(= (list 1) (sorted (list 1)))`,
		`(= (list 1 2 3) (append nil (list 1 2 3)))`,
		`(let ((result (fold-left (lambda (x y) (if (> (* x y) 10) 10 (* x y))) 1 (filter (lambda (x) maybe) (list 2 3 6 7 8 9 10 11 12 13 14)))))
		   (and (_maybe? (= 2 result))
			      (not (= 5 result))))`,
		`(let ((result (fold-left (lambda (x y) (if (> (* x y) 10) 10 (* x y))) 1 (filter (lambda (x) maybe) (list 2 2 3 6 7 8 9 10 11 12 13 14)))))
		   (and (_maybe? (= 2 result))
			      (not (= 5 result))))`,
		`(= "(0 (1 (2 (3 (4 nil)))))" (_to-string (fold-right list nil (range 5))))`,
		`(= "(((((nil 0) 1) 2) 3) 4)" (_to-string (fold-left list nil (range 5))))`,
		`(contains-duplicates? (list 1 2 3 4 5 2))`,
		`(not (contains-duplicates? (list 1 2 3 4 5)))`,
		`(_maybe? (contains-duplicates? (list 1 2 3 4 5 (any-of 5 6))))`,
		`(not (contains-duplicates? (list 1 2 3 4 5 (any-of 6 7))))`,
		`(_maybe? (contains-duplicates? (list 1 2 3 4 5 (any-of 6 7) (any-of 7 8))))`,
		`(_maybe? (contains-duplicates? (list 1 2 3 4 5 unknown)))`,
	}

	onExpression := func(category string, i int, s string, aspirational bool) {
		rv, err := parser.Parse(fmt.Sprintf("<testcase %d>", i), []byte(s))
		if err != nil {
			t.Errorf("error parsing %s #%d %q: %v", category, i, s, err)
			return
		}

		exprs, ok := rv.([]interface{})
		if !ok {
			t.Errorf("error parsing %s #%d %q: %v", category, i, s, err)
		}

		var result types.Value
		for j, xpr := range exprs {
			result, err = xpr.(types.Value).Eval(root)
			if err != nil {
				if !aspirational {
					t.Errorf("error evaluating %s #%d %q (sub-expression %d: %v): %v", category, i, s, j, xpr, err)
				}
				break
			}
		}
		if err != nil {
			return
		}

		if unknown.IsUncertain(result) {
			if !aspirational {
				t.Errorf("uncertain result for %s #%d %q: %v", category, i, s, result)
			}
			return
		}

		if result.Falsey() {
			if !aspirational {
				t.Errorf("falsey result for %s #%d %q: %v", category, i, s, result)
			}
			return
		}

		if aspirational {
			log.Printf("aspirational expression %q is now true!", s)
		}
	}

	for i, s := range exprs {
		onExpression("main", i, s, false)
	}

	// These things are for some reason not easy to achieve (without coding
	// specifically for them).
	thingsWeWouldLikeToBeTrue := []string{
		`(contains-duplicates? (list 1 2 3 4 5 (any-of 4 5)))`,
		`(letfunc* ((my-fact (n) (if (= 0 n) 1 (* n (my-fact (- n 1)))))) (= (my-fact 5) 120))`,
	}

	for i, s := range thingsWeWouldLikeToBeTrue {
		onExpression("aspirational", i, s, true)
	}
}

var values = []string{
	"123",
	"3.14",
	"true",
	"false",
	"maybe",
	"unknown",
	"\"hello\"",
	"(quote hello)",
	"(list 1 2 3)",
	"(list 1 2 4)",
	"(cons 1 2)",
	"(cons 1 3)",
	"(list)",
	"(any-of 123 124)",
	"(list 1 2 (any-of 3 4))",
	"(any-of 1.5 2)",
	"(list 1.5 2)",
	"(unknown-of-type 'integer)",
	"(number-in-range 'from 1 'to 42)",
	"NaN",
	"1/2",
	"36893488147419103232",
}

func TestUnaryInvariants(t *testing.T) {
	templates := []string{
		"(or (must? (nan? EXPR)) (may? (equals? EXPR EXPR)))",
		"(or (uncertain? EXPR) (must? (nan? EXPR)) (equals? EXPR EXPR))",
	}

	env := builtin.NewRootEnv()

	for _, inserted := range values {
		for _, template := range templates {
			s := strings.Replace(template, "EXPR", inserted, -1)
			name := fmt.Sprintf("<unary invariant: %q>", s)

			result, err := code.Run(env, name, []byte(s))
			if err != nil {
				t.Errorf("code.Run(..., %q) = err: %v", s, err)
				continue
			}

			if unknown.IsUncertain(result) {
				t.Errorf("code.Run(..., %q) = %v (uncertain)", s, result)
				continue
			}

			if result.Falsey() {
				t.Errorf("code.Run(..., %q) = %v (falsey)", s, result)
			}
		}
	}
}

func TestBinaryInvariants(t *testing.T) {
	templates := []string{
		"(_dumb-equals? (= EXPR1 EXPR2) (= EXPR2 EXPR1))",
	}

	env := builtin.NewRootEnv()

	for _, inserted1 := range values {
		for _, inserted2 := range values {
			for _, template := range templates {
				s := strings.Replace(template, "EXPR1", inserted1, -1)
				s = strings.Replace(s, "EXPR2", inserted2, -1)

				name := fmt.Sprintf("<binary invariant: %q>", s)

				result, err := code.Run(env, name, []byte(s))
				if err != nil {
					t.Errorf("code.Run(..., %q) = err: %v", s, err)
					continue
				}

				if unknown.IsUncertain(result) {
					t.Errorf("code.Run(..., %q) = %v (uncertain)", s, result)
					continue
				}

				if result.Falsey() {
					t.Errorf("code.Run(..., %q) = %v (falsey)", s, result)
				}
			}
		}
	}
}

func listLispFilesInOrder(dirname string) ([]string, error) {
	infos, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	var rv []string
	for _, info := range infos {
		if strings.HasSuffix(info.Name(), ".hlisp") {
			rv = append(rv, filepath.Join(dirname, info.Name()))
		}
	}
	sort.Strings(rv)
	return rv, nil
}

func TestExamples(t *testing.T) {
	dirname := "./examples/"

	worksInProgress := []string{}

	filenames, err := listLispFilesInOrder(dirname)
	if err != nil {
		t.Fatalf("listLispFilesInOrder(%q) = err: %v", dirname, err)
	}

	wip := map[string]bool{}
	for _, s := range worksInProgress {
		wip[s] = true
	}

	for _, filename := range filenames {
		isWIP := wip[filepath.Base(filename)]

		_, err := code.RunFile(builtin.NewRootEnv(), filename)
		if !isWIP {
			if err != nil {
				t.Errorf("code.RunFile(..., %q) = err: %v", filename, err)
			}
		} else {
			status := "now passing :D"
			if err != nil {
				status = "still failing "
			}
			log.Printf("WIP test %s %q: %v", status, filename, err)
		}
	}
}

func BenchmarkTuringBinaryCounter20(b *testing.B) {
	turingSetupCode := []byte(`
(defun! remove-top (tape)
  (if (nil? tape)
      nil
      (cdr tape)))

(defun! get-top (tape)
  (if (nil? tape)
      nil
      (car tape)))

(defun! move-left (left-tape current right-tape)
  (list (remove-top left-tape)
        (get-top left-tape)
        (cons current right-tape)))

(defun! move-right (left-tape current right-tape)
  (list (cons current left-tape)
        (get-top right-tape)
        (remove-top right-tape)))

(set! do-not-move list)

(defun! compute-turing (left-tape current right-tape state table-func)
  (let ((recurse (lambda (new-symbol movement new-state)
                    (let ((current new-symbol))
                      (let ((next-tape-states (movement left-tape current right-tape)))
                        (if (_atom-eq? new-state 'stop)
                            next-tape-states
                            (let ((left-tape (first next-tape-states))
                                  (current (second next-tape-states))
                                  (right-tape (third next-tape-states)))
                              (compute-turing left-tape current right-tape new-state table-func))))))))
    (let ((decision (table-func state current)))
      (let ((new-symbol (first decision))
            (move-func (second decision))
            (new-state (third decision)))
        (recurse new-symbol move-func new-state)))))

(defun! binary-counter-table (state sym)
  (cond
    ((and (= state 0) (= sym nil)) (list nil move-right 1))
    ((and (= state 0) (= sym 0))   (list 0 move-left 0))
    ((and (= state 0) (= sym 1))   (list 1 move-left 0))

    ((and (= state 1) (= sym nil)) (list 1 move-left 2))
    ((and (= state 1) (= sym 0))   (list 1 move-right 2))
    ((and (= state 1) (= sym 1))   (list 0 move-right 1))

    ((and (= state 2) (= sym nil)) (list nil move-right 'stop))
    ((and (= state 2) (= sym 0))   (list 0 move-left 2))
    ((and (= state 2) (= sym 1))   (list 1 move-left 2))
    (true (call-nonexistent-function-to-fail))))

(defun! turing-inc (xs) (compute-turing (first xs) (second xs) (third xs) 0 binary-counter-table))

(defun! turing-inc-n (n xs) (if (= 0 n) xs (turing-inc-n (dec n) (turing-inc xs))))

(defun! align-tape (xs)
  (if (and (nil? (second xs)) (nil? (third xs)))
      (first xs)
      (align-tape (move-right (first xs) (second xs) (third xs)))))

(defun! turing-integer-to-binary (n) (align-tape (turing-inc-n n (list nil nil nil))))
`)
	root := builtin.NewRootEnv()
	_, err := code.Run(root, "<benchmark setup code>", turingSetupCode)
	if err != nil {
		b.Fatalf("failed to set up benchmark code: %v", err)
	}

	turingCode := []byte("(turing-integer-to-binary 20)")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := code.Run(root, "<turing benchmark code>", turingCode)
		if err != nil {
			b.Fatalf("evaluating benchmark code %q failed: %v", turingCode, err)
		}
	}
}

func BenchmarkGoBinaryCounter20(b *testing.B) {
	reversed := func(xs []int) []int {
		var rv []int
		for i := len(xs) - 1; i >= 0; i-- {
			rv = append(rv, xs[i])
		}
		return rv
	}

	var binaryIncReversed func(xs []int) []int

	binaryIncReversed = func(xs []int) []int {
		if len(xs) == 0 {
			return []int{1}
		}
		if xs[0] == 0 {
			return append([]int{1}, xs[1:]...)
		}
		return append([]int{0}, binaryIncReversed(xs[1:])...)
	}

	intToBinary := func(n int) []int {
		x := []int{0}
		for i := 0; i < n; i++ {
			x = binaryIncReversed(x)
		}
		return reversed(x)
	}

	got := intToBinary(1000)
	want := []int{1, 1, 1, 1, 1, 0, 1, 0, 0, 0}

	if !reflect.DeepEqual(got, want) {
		b.Fatalf("intToBinary(1000) = %v want %v", got, want)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got = intToBinary(20)
	}
}

func BenchmarkNormalBinaryCounter20(b *testing.B) {
	counterSetupCode := []byte(`
(defun! binary-inc (xs)
  (reversed (binary-inc-reversed (reversed xs))))

(defun! binary-inc-reversed (xs)
  (if (empty? xs)
      (list 1)
      (if (= 0 (car xs))
          (cons 1 (cdr xs))
          (cons 0 (binary-inc-reversed (cdr xs))))))

(defun! unary-n-times (n f x)
  (if (= n 0) x (unary-n-times (- n 1) f (f x))))

(defun! simple-integer-to-binary (n)
  (reversed (unary-n-times n binary-inc-reversed (list 0))))
`)
	root := builtin.NewRootEnv()
	_, err := code.Run(root, "<benchmark setup code>", counterSetupCode)
	if err != nil {
		b.Fatalf("failed to set up benchmark code: %v", err)
	}

	turingCode := []byte("(simple-integer-to-binary 20)")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := code.Run(root, "<turing benchmark code>", turingCode)
		if err != nil {
			b.Fatalf("evaluating benchmark code %q failed: %v", turingCode, err)
		}
	}
}

func BenchmarkSortList50(b *testing.B) {
	sortSetupCode := []byte(`
(set! my-list (append (range 50) (range 50)))
`)
	root := builtin.NewRootEnv()
	_, err := code.Run(root, "<benchmark setup code>", sortSetupCode)
	if err != nil {
		b.Fatalf("failed to set up benchmark code: %v", err)
	}

	sortCode := []byte("(sorted my-list)")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := code.Run(root, "<sort benchmark code>", sortCode)
		if err != nil {
			b.Fatalf("evaluating benchmark code %q failed: %v", sortCode, err)
		}
	}
}

func BenchmarkSortList200(b *testing.B) {
	sortSetupCode := []byte(`
(set! my-list (append (range 200) (range 200)))
`)
	root := builtin.NewRootEnv()
	_, err := code.Run(root, "<benchmark setup code>", sortSetupCode)
	if err != nil {
		b.Fatalf("failed to set up benchmark code: %v", err)
	}

	sortCode := []byte("(sorted my-list)")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := code.Run(root, "<sort benchmark code>", sortCode)
		if err != nil {
			b.Fatalf("evaluating benchmark code %q failed: %v", sortCode, err)
		}
	}
}

func BenchmarkLengthOfListList500(b *testing.B) {
	lengthSetupCode := []byte(`
(set! my-list (range 500))
`)
	root := builtin.NewRootEnv()
	_, err := code.Run(root, "<benchmark setup code>", lengthSetupCode)
	if err != nil {
		b.Fatalf("failed to set up benchmark code: %v", err)
	}

	lengthCode := []byte("(length my-list)")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := code.Run(root, "<length benchmark code>", lengthCode)
		if err != nil {
			b.Fatalf("evaluating benchmark code %q failed: %v", lengthCode, err)
		}
	}
}

func BenchmarkNormalConsFoldLeft500(b *testing.B) {
	root := builtin.NewRootEnv()

	ocflCode := []byte("(fold-left (lambda (x y) (mod (* x y) 10)) 1 (range 500))")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := code.Run(root, "<normalcons-fold-left benchmark code>", ocflCode)
		if err != nil {
			b.Fatalf("evaluating benchmark code %q failed: %v", ocflCode, err)
		}
	}
}

func BenchmarkNormalConsFoldRight500(b *testing.B) {
	root := builtin.NewRootEnv()

	ocflCode := []byte("(fold-right (lambda (x y) (mod (* x y) 10)) 1 (range 500))")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := code.Run(root, "<normalcons-fold-right benchmark code>", ocflCode)
		if err != nil {
			b.Fatalf("evaluating benchmark code %q failed: %v", ocflCode, err)
		}
	}
}

func BenchmarkOptConsFoldLeft50(b *testing.B) {
	root := builtin.NewRootEnv()

	ocflCode := []byte("(fold-left (lambda (x y) (mod (* x y) 10)) 1 (filter (lambda (x) maybe) (range 50)))")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := code.Run(root, "<optcons-fold-left benchmark code>", ocflCode)
		if err != nil {
			b.Fatalf("evaluating benchmark code %q failed: %v", ocflCode, err)
		}
	}
}

func BenchmarkOptConsFoldRight50(b *testing.B) {
	root := builtin.NewRootEnv()

	ocflCode := []byte("(fold-right (lambda (x y) (mod (* x y) 10)) 1 (filter (lambda (x) maybe) (range 50)))")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := code.Run(root, "<optcons-fold-right benchmark code>", ocflCode)
		if err != nil {
			b.Fatalf("evaluating benchmark code %q failed: %v", ocflCode, err)
		}
	}
}
