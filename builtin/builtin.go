package builtin

import (
	"errors"
	"fmt"

	"github.com/steinarvk/heisenlisp/env"
	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/function"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/unknown"
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

func Values(e types.Env, name string, f func(xs []types.Value) (types.Value, error)) {
	e.Bind(name, expr.NewBuiltinFunction(name, f))
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

	tv, err := unknown.TruthValue(condition)
	if err != nil {
		return nil, err
	}

	switch tv {
	case unknown.Maybe:
		thenVal, err := thenClause.Eval(e)
		if err != nil {
			return nil, err
		}
		elseVal, err := elseClause.Eval(e)
		if err != nil {
			return nil, err
		}

		return unknown.NewMaybeAnyOf([]types.Value{
			thenVal, elseVal,
		}), nil
	case unknown.True:
		return thenClause.Eval(e)
	case unknown.False:
		return elseClause.Eval(e)
	}
	return nil, errors.New("impossible state: ternary truth value neither true, false, or maybe")
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

type quasiquoteSpecialForm struct{}

func (i quasiquoteSpecialForm) TypeName() string                    { return "special" }
func (i quasiquoteSpecialForm) String() string                      { return specialFormString("quasiquote") }
func (i quasiquoteSpecialForm) Falsey() bool                        { return false }
func (i quasiquoteSpecialForm) Uncertain() bool                     { return false }
func (i quasiquoteSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i quasiquoteSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) != 1 {
		return nil, fmt.Errorf("quasiquote: unary special form, got %d params", len(unevaluated))
	}
	return expr.ExpandQuasiQuote(e, unevaluated[0])
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

type defmacroSpecialForm struct{}

func (i defmacroSpecialForm) TypeName() string                    { return "special" }
func (i defmacroSpecialForm) String() string                      { return specialFormString("defmacro!") }
func (i defmacroSpecialForm) Falsey() bool                        { return false }
func (i defmacroSpecialForm) Uncertain() bool                     { return false }
func (i defmacroSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i defmacroSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) < 3 {
		return nil, fmt.Errorf("defmacro!: too few arguments")
	}

	name, err := expr.SymbolName(unevaluated[0])
	if err != nil {
		return nil, fmt.Errorf("defmacro! name: %v", err)
	}

	macroValue, err := function.NewMacro(e, name, unevaluated[1], unevaluated[2:])
	if err != nil {
		return nil, err
	}

	e.Bind(name, macroValue)

	return macroValue, nil
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

type andSpecialForm struct{}

func (i andSpecialForm) TypeName() string                    { return "special" }
func (i andSpecialForm) String() string                      { return specialFormString("and") }
func (i andSpecialForm) Falsey() bool                        { return false }
func (i andSpecialForm) Uncertain() bool                     { return false }
func (i andSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i andSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	knownToMaybeBeFalse := false

	for _, uneval := range unevaluated {
		eval, err := uneval.Eval(e)
		if err != nil {
			return nil, err
		}
		truth, err := unknown.TruthValue(eval)
		if err != nil {
			return nil, err
		}
		switch truth {
		case unknown.True:
			break
		case unknown.False:
			return expr.Bool(false), nil
		default:
			knownToMaybeBeFalse = true
		}
	}

	if knownToMaybeBeFalse {
		return unknown.NewMaybeAnyOf([]types.Value{
			expr.Bool(true),
			expr.Bool(false),
		}), nil
	}

	return expr.Bool(true), nil
}

type orSpecialForm struct{}

func (i orSpecialForm) TypeName() string                    { return "special" }
func (i orSpecialForm) String() string                      { return specialFormString("or") }
func (i orSpecialForm) Falsey() bool                        { return false }
func (i orSpecialForm) Uncertain() bool                     { return false }
func (i orSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i orSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	knownToMaybeBeTrue := false

	for _, uneval := range unevaluated {
		eval, err := uneval.Eval(e)
		if err != nil {
			return nil, err
		}
		truth, err := unknown.TruthValue(eval)
		if err != nil {
			return nil, err
		}
		switch truth {
		case unknown.True:
			return expr.Bool(true), nil
		case unknown.False:
			break
		default:
			knownToMaybeBeTrue = true
		}
	}

	if knownToMaybeBeTrue {
		return unknown.NewMaybeAnyOf([]types.Value{
			expr.Bool(true),
			expr.Bool(false),
		}), nil
	}

	return expr.Bool(false), nil
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
	name := ""
	if namePtr != nil {
		name = *namePtr
	}

	return function.New(lexicalEnv, name, formalParamSpec, body)
}

func BindDefaults(e types.Env) {
	e.Bind("if", &ifSpecialForm{})
	e.Bind("set!", &setSpecialForm{})
	e.Bind("defun!", &defunSpecialForm{})
	e.Bind("defmacro!", &defmacroSpecialForm{})
	e.Bind("lambda", &lambdaSpecialForm{})
	e.Bind("quote", &quoteSpecialForm{})
	e.Bind("quasiquote", &quasiquoteSpecialForm{})
	e.Bind("let", &letSpecialForm{})

	e.Bind("nil", expr.NilValue{})
	e.Bind("true", expr.Bool(true))
	e.Bind("false", expr.Bool(false))

	Unary(e, "_atom?", func(a types.Value) (types.Value, error) {
		_, ok := a.(types.Atom)
		return expr.Bool(ok), nil
	})

	Binary(e, "_atom-eq?", func(a, b types.Value) (types.Value, error) {
		return expr.Bool(expr.AtomEquals(a, b)), nil
	})

	Unary(e, "_type", func(a types.Value) (types.Value, error) {
		return expr.ToSymbol(a.TypeName()), nil
	})

	Unary(e, "_unknown?", func(a types.Value) (types.Value, error) {
		return expr.Bool(a.Uncertain()), nil
	})

	Binary(e, "cons", func(a, b types.Value) (types.Value, error) {
		return expr.Cons(a, b), nil
	})

	Unary(e, "car", func(a types.Value) (types.Value, error) {
		return expr.Car(a)
	})

	Unary(e, "cdr", func(a types.Value) (types.Value, error) {
		return expr.Cdr(a)
	})

	Unary(e, "not", func(a types.Value) (types.Value, error) {
		val, err := unknown.TruthValue(a)
		if err != nil {
			return nil, err
		}
		switch val {
		case unknown.False:
			return expr.Bool(true), nil
		case unknown.True:
			return expr.Bool(false), nil
		case unknown.Maybe:
			return unknown.NewMaybeAnyOf([]types.Value{
				expr.Bool(true), expr.Bool(false),
			}), nil
		}
		return nil, errors.New("impossible state: ternary truth value neither true, false, or maybe")
	})

	Values(e, "any-of", func(xs []types.Value) (types.Value, error) {
		return unknown.NewMaybeAnyOf(xs), nil
	})

	Unary(e, "possible-values", func(v types.Value) (types.Value, error) {
		vals, ok := unknown.PossibleValues(v)
		if ok {
			return expr.WrapList(vals), nil
		}
		return unknown.FullyUnknown{}, nil
	})

	// todo: the arithmetic functions need to be made unknown-aware.

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

	Integers(e, "mod", func(xs []expr.Integer) (types.Value, error) {
		if len(xs) != 2 {
			return nil, fmt.Errorf("mod: got %d params want 2", len(xs))
		}
		if xs[1] == 0 {
			return nil, errors.New("division by zero")
		}
		return expr.Integer(int64(xs[0]) % int64(xs[1])), nil
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

func NewRootEnv() types.Env {
	rv := env.New(nil)
	BindDefaults(rv)
	return rv
}
