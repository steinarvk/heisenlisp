package builtin

import (
	"errors"
	"fmt"

	"github.com/steinarvk/heisenlisp/env"
	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/types"
)

func wrap(name string, f func(a []types.Value) (types.Value, error)) func([]types.Value) (types.Value, error) {
	fw := func(vs []types.Value) (types.Value, error) {
		rv, err := f(vs)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", name, err)
		}
		return rv, nil
	}
	return fw
}

func Nullary(e types.Env, name string, f func() (types.Value, error)) {
	checker := func(vs []types.Value) (types.Value, error) {
		if len(vs) != 0 {
			return nil, fmt.Errorf("want no params, got %d", len(vs))
		}
		return f()
	}
	e.Bind(name, expr.NewBuiltinFunction(name, wrap(name, checker)))
}

func Unary(e types.Env, name string, f func(a types.Value) (types.Value, error)) {
	checker := func(vs []types.Value) (types.Value, error) {
		if len(vs) != 1 {
			return nil, fmt.Errorf("want 1 param, got %d", len(vs))
		}
		return f(vs[0])
	}
	e.Bind(name, expr.NewBuiltinFunction(name, wrap(name, checker)))
}

func Binary(e types.Env, name string, f func(a, b types.Value) (types.Value, error)) {
	checker := func(vs []types.Value) (types.Value, error) {
		if len(vs) != 2 {
			return nil, fmt.Errorf("want 2 params, got %d", len(vs))
		}
		return f(vs[0], vs[1])
	}
	e.Bind(name, expr.NewBuiltinFunction(name, wrap(name, checker)))
}

func Integers(e types.Env, name string, f func([]expr.Integer) (types.Value, error)) {
	checker := func(vs []types.Value) (types.Value, error) {
		var iparams []expr.Integer
		for i, v := range vs {
			iv, ok := v.(expr.Integer)
			if !ok {
				return nil, fmt.Errorf("param #%d: want Integer, got %v", i, v)
			}
			iparams = append(iparams, iv)
		}
		return f(iparams)
	}
	e.Bind(name, expr.NewBuiltinFunction(name, wrap(name, checker)))
}

func specialFormString(s string) string { return fmt.Sprintf("#<special %q>", s) }

type ifSpecialForm struct{}

func (i ifSpecialForm) TypeName() string                    { return "special" }
func (i ifSpecialForm) String() string                      { return specialFormString("if") }
func (i ifSpecialForm) Falsey() bool                        { return false }
func (i ifSpecialForm) Uncertain() bool                     { return false }
func (i ifSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i ifSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) != 3 {
		return nil, fmt.Errorf("'if' expects 3 params, got %d", len(unevaluated))
	}
	conditionClause := unevaluated[0]
	thenClause := unevaluated[1]
	elseClause := unevaluated[2]

	condition, err := conditionClause.Eval(e)
	if err != nil {
		return nil, err
	}
	if condition.Uncertain() {
		return nil, fmt.Errorf("branching on uncertain value: %v", condition)
	}

	if condition.Falsey() {
		return elseClause.Eval(e)
	}
	return thenClause.Eval(e)
}

type setSpecialForm struct{}

func (i setSpecialForm) TypeName() string                    { return "special" }
func (i setSpecialForm) String() string                      { return specialFormString("set!") }
func (i setSpecialForm) Falsey() bool                        { return false }
func (i setSpecialForm) Uncertain() bool                     { return false }
func (i setSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i setSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) != 2 {
		return nil, fmt.Errorf("'set!' expects 2 params, got %d", len(unevaluated))
	}

	symbol, ok := unevaluated[0].(expr.Identifier)
	if !ok {
		return nil, fmt.Errorf("must (set!) symbol, not %v", unevaluated[0])
	}

	value, err := unevaluated[1].Eval(e)
	if err != nil {
		return nil, err
	}

	e.Bind(string(symbol), value)

	return value, nil
}

type quoteSpecialForm struct{}

func (i quoteSpecialForm) TypeName() string                    { return "special" }
func (i quoteSpecialForm) String() string                      { return specialFormString("quote") }
func (i quoteSpecialForm) Falsey() bool                        { return false }
func (i quoteSpecialForm) Uncertain() bool                     { return false }
func (i quoteSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i quoteSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) != 1 {
		return nil, fmt.Errorf("quote: unary special form, got %d params", len(unevaluated))
	}
	return unevaluated[0], nil
}

type defunSpecialForm struct{}

func (i defunSpecialForm) TypeName() string                    { return "special" }
func (i defunSpecialForm) String() string                      { return specialFormString("defun!") }
func (i defunSpecialForm) Falsey() bool                        { return false }
func (i defunSpecialForm) Uncertain() bool                     { return false }
func (i defunSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i defunSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) < 3 {
		return nil, fmt.Errorf("defun!: too few arguments")
	}

	nameSym, ok := unevaluated[0].(expr.Identifier)
	if !ok {
		return nil, fmt.Errorf("must defun! symbol as name, not %v", unevaluated[0])
	}

	name := string(nameSym)
	funcVal, err := makeFunction(&name, e, unevaluated[1], unevaluated[2:])
	if err != nil {
		return nil, err
	}

	e.Bind(name, funcVal)
	return funcVal, nil
}

type letSpecialForm struct{}

func (i letSpecialForm) TypeName() string                    { return "special" }
func (i letSpecialForm) String() string                      { return specialFormString("let") }
func (i letSpecialForm) Falsey() bool                        { return false }
func (i letSpecialForm) Uncertain() bool                     { return false }
func (i letSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i letSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	// (let (bindings) forms...)
	if len(unevaluated) < 2 {
		return nil, fmt.Errorf("let: too few arguments")
	}

	bindings, err := expr.UnwrapList(unevaluated[0])
	if err != nil {
		return nil, fmt.Errorf("error unwrapping bindings: %v", err)
	}

	childEnv := env.New(e)

	for i, binding := range bindings {
		bindingList, err := expr.UnwrapList(binding)
		if err != nil {
			return nil, fmt.Errorf("binding %d: error unwrapping: %v", i, err)
		}
		if len(bindingList) != 2 {
			return nil, fmt.Errorf("binding %d: wrong length (want 2): %d", i, len(bindingList))
		}

		sym, err := expr.SymbolName(bindingList[0])
		if err != nil {
			return nil, fmt.Errorf("binding %d: error getting binding name: %v", i, err)
		}

		val, err := bindingList[1].Eval(e)
		if err != nil {
			return nil, err
		}

		childEnv.Bind(sym, val)
	}

	return expr.Progn(childEnv, unevaluated[1:])
}

type lambdaSpecialForm struct{}

func (i lambdaSpecialForm) TypeName() string                    { return "special" }
func (i lambdaSpecialForm) String() string                      { return specialFormString("lambda") }
func (i lambdaSpecialForm) Falsey() bool                        { return false }
func (i lambdaSpecialForm) Uncertain() bool                     { return false }
func (i lambdaSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i lambdaSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) < 2 {
		return nil, fmt.Errorf("lambda: too few arguments")
	}

	funcVal, err := makeFunction(nil, e, unevaluated[0], unevaluated[1:])
	if err != nil {
		return nil, err
	}

	return funcVal, nil
}

func makeFunction(namePtr *string, lexicalEnv types.Env, formalParamSpec types.Value, body []types.Value) (types.Value, error) {
	formalParams, err := expr.UnwrapList(formalParamSpec)
	if err != nil {
		return nil, fmt.Errorf("params specifier must be list, not %v: %v", formalParamSpec, err)
	}

	var formalParamNames []string

	for _, formalParam := range formalParams {
		s, ok := formalParam.(expr.Identifier)
		if !ok {
			return nil, fmt.Errorf("formal param spec must be symbol, not %v", formalParam)
		}
		formalParamNames = append(formalParamNames, string(s))
	}

	name := ""
	if namePtr != nil {
		name = *namePtr
	}

	funcVal := expr.NewLispFunction(lexicalEnv, name, formalParamNames, body)

	return funcVal, nil
}

func BindDefaults(e types.Env) {
	e.Bind("if", &ifSpecialForm{})
	e.Bind("set!", &setSpecialForm{})
	e.Bind("defun!", &defunSpecialForm{})
	e.Bind("lambda", &lambdaSpecialForm{})
	e.Bind("quote", &quoteSpecialForm{})
	e.Bind("let", &letSpecialForm{})

	Unary(e, "_atom?", func(a types.Value) (types.Value, error) {
		_, ok := a.(types.Atom)
		return expr.Bool(ok), nil
	})

	Binary(e, "_atom-eq?", func(a, b types.Value) (types.Value, error) {
		av, aok := a.(types.Atom)
		bv, bok := b.(types.Atom)
		return expr.Bool(aok && bok && av.AtomEquals(bv)), nil
	})

	Unary(e, "_type", func(a types.Value) (types.Value, error) {
		return expr.ToSymbol(a.TypeName()), nil
	})

	Binary(e, "cons", func(a, b types.Value) (types.Value, error) {
		return &expr.ConsValue{a, b}, nil
	})

	Unary(e, "car", func(a types.Value) (types.Value, error) {
		ca, ok := a.(*expr.ConsValue)
		if !ok {
			return nil, errors.New("not a cons")
		}
		return ca.Car, nil
	})

	Unary(e, "cdr", func(a types.Value) (types.Value, error) {
		ca, ok := a.(*expr.ConsValue)
		if !ok {
			return nil, errors.New("not a cons")
		}
		return ca.Cdr, nil
	})

	// convenience bindings
	e.Bind("true", expr.Bool(true))
	e.Bind("false", expr.Bool(false))
	e.Bind("nil", expr.NilValue{})

	Unary(e, "not", func(a types.Value) (types.Value, error) {
		if a.Uncertain() {
			return nil, fmt.Errorf("TODO: negating uncertain values not yet implemented")
		}

		return expr.Bool(a.Falsey()), nil
	})

	Integers(e, "+", func(xs []expr.Integer) (types.Value, error) {
		var rv int64
		for _, x := range xs {
			rv += int64(x)
		}
		return expr.Integer(rv), nil
	})

	Integers(e, "-", func(xs []expr.Integer) (types.Value, error) {
		switch {
		case len(xs) == 0:
			return expr.Integer(0), nil
		case len(xs) == 1:
			return expr.Integer(-xs[0]), nil
		default:
			rv := int64(xs[0])
			for _, x := range xs[1:] {
				rv -= int64(x)
			}
			return expr.Integer(rv), nil
		}
	})

	Integers(e, "*", func(xs []expr.Integer) (types.Value, error) {
		rv := int64(1)
		for _, x := range xs {
			rv *= int64(x)
		}
		return expr.Integer(rv), nil
	})

	Integers(e, "=", func(xs []expr.Integer) (types.Value, error) {
		if len(xs) == 0 {
			return nil, fmt.Errorf("cannot check equality of no values")
		}
		v := int64(xs[0])
		for _, x := range xs[1:] {
			if int64(x) != v {
				return expr.Bool(false), nil
			}
		}
		return expr.Bool(true), nil
	})
}
