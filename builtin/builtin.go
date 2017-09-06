package builtin

import (
	"fmt"

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
	e.Bind(name, expr.NewFunction(e, name, wrap(name, checker)))
}

func Unary(e types.Env, name string, f func(a types.Value) (types.Value, error)) {
	checker := func(vs []types.Value) (types.Value, error) {
		if len(vs) != 1 {
			return nil, fmt.Errorf("want 1 param, got %d", len(vs))
		}
		return f(vs[0])
	}
	e.Bind(name, expr.NewFunction(e, name, wrap(name, checker)))
}

func Binary(e types.Env, name string, f func(a, b types.Value) (types.Value, error)) {
	checker := func(vs []types.Value) (types.Value, error) {
		if len(vs) != 2 {
			return nil, fmt.Errorf("want 2 params, got %d", len(vs))
		}
		return f(vs[0], vs[1])
	}
	e.Bind(name, expr.NewFunction(e, name, wrap(name, checker)))
}

func specialFormString(s string) string { return fmt.Sprintf("#<special %q>", s) }

type ifSpecialForm struct{}

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

func BindDefaults(e types.Env) {
	e.Bind("if", &ifSpecialForm{})
	e.Bind("set!", &setSpecialForm{})
}
