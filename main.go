package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/env"
	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/types"

	"github.com/steinarvk/heisenlisp/gen/parser"
)

func mainCore() error {
	wr := bufio.NewWriter(os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)

	prompt := "..? "

	root := env.New(nil)
	builtin.BindDefaults(root)

	builtin.Nullary(root, "random!", func() (types.Value, error) {
		return expr.Integer(rand.Intn(1000)), nil
	})

	builtin.Binary(root, "+", func(a, b types.Value) (types.Value, error) {
		an, ok := a.(expr.Integer)
		if !ok {
			return nil, fmt.Errorf("not an integer: %v", a)
		}

		bn, ok := b.(expr.Integer)
		if !ok {
			return nil, fmt.Errorf("not an integer: %v", b)
		}

		return expr.Integer(int64(an) + int64(bn)), nil
	})

	builtin.Binary(root, "-", func(a, b types.Value) (types.Value, error) {
		an, ok := a.(expr.Integer)
		if !ok {
			return nil, fmt.Errorf("not an integer: %v", a)
		}

		bn, ok := b.(expr.Integer)
		if !ok {
			return nil, fmt.Errorf("not an integer: %v", b)
		}

		return expr.Integer(int64(an) - int64(bn)), nil
	})

	builtin.Binary(root, "*", func(a, b types.Value) (types.Value, error) {
		an, ok := a.(expr.Integer)
		if !ok {
			return nil, fmt.Errorf("not an integer: %v", a)
		}

		bn, ok := b.(expr.Integer)
		if !ok {
			return nil, fmt.Errorf("not an integer: %v", b)
		}

		return expr.Integer(int64(an) * int64(bn)), nil
	})

	wr.Write([]byte(prompt))
	wr.Flush()

	for scanner.Scan() {
		text := scanner.Text()

		if strings.TrimSpace(text) == "" {
			wr.Write([]byte(prompt))
			wr.Flush()
			continue
		}

		rv, err := parser.Parse("<stdin>", []byte(text))
		if err != nil {
			wr.Write([]byte(fmt.Sprintf("==! parsing error: %v\n", err)))
		} else {
			wr.Write([]byte(fmt.Sprintf("(read) ==> %v\n", rv)))

			evaled, err := rv.(types.Value).Eval(root)
			if err != nil {
				wr.Write([]byte(fmt.Sprintf("==! eval error: %v\n", err)))
			} else {
				wr.Write([]byte(fmt.Sprintf("(eval) ==> %v\n", evaled)))
			}
		}

		wr.Write([]byte(prompt))
		wr.Flush()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}

	return nil
}

func main() {
	if err := mainCore(); err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
	}
}
