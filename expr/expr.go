package expr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/steinarvk/heisenlisp/types"
)

type Bool bool

func (b Bool) String() string {
	if bool(b) {
		return "#true"
	}
	return "#false"
}

func (b Bool) Eval(_ types.Env) (types.Value, error) { return b, nil }

type AnyOf []types.Value

func (a AnyOf) String() string {
	var xs []string
	for _, x := range a {
		xs = append(xs, x.String())
	}
	return fmt.Sprintf("#any-of(%s)", strings.Join(xs, " "))
}

func (a AnyOf) Eval(_ types.Env) (types.Value, error) { return a, nil }

type FullyUnknown struct{}

func (_ FullyUnknown) String() string                        { return "#unknown" }
func (f FullyUnknown) Eval(_ types.Env) (types.Value, error) { return f, nil }

type Identifier string

func (i Identifier) String() string {
	return string(i)
}

func (i Identifier) Eval(e types.Env) (types.Value, error) {
	val, ok := e.Lookup(string(i))
	if !ok {
		return nil, fmt.Errorf("no such identifier: %q", i)
	}
	return val, nil
}

type Integer int64

func (i Integer) String() string {
	return fmt.Sprintf("%d", i)
}

func (i Integer) Eval(_ types.Env) (types.Value, error) { return i, nil }

type String string

func (s String) String() string {
	return fmt.Sprintf("%q", string(s))
}

func (s String) Eval(_ types.Env) (types.Value, error) { return s, nil }

type ListValue []types.Value

func (l ListValue) String() string {
	var xs []string
	for _, x := range l {
		xs = append(xs, x.String())
	}
	return fmt.Sprintf("(%s)", strings.Join(xs, " "))
}

func (l ListValue) Eval(e types.Env) (types.Value, error) {
	if len(l) < 1 {
		return nil, errors.New("cannot evaluate empty list")
	}
	funcVal, err := l[0].Eval(e)
	if err != nil {
		return nil, err
	}

	callable, ok := funcVal.(types.Callable)
	if !ok {
		return nil, fmt.Errorf("%q (%v) is not callable", l[0], funcVal)
	}

	var params []types.Value
	for _, unevaled := range l[1:] {
		evaled, err := unevaled.Eval(e)
		if err != nil {
			return nil, err
		}
		params = append(params, evaled)
	}

	return callable.Call(params)
}

type FunctionValue struct {
	name       string
	lexicalEnv types.Env
	function   func([]types.Value) (types.Value, error)
}

func NewFunction(env types.Env, name string, f func([]types.Value) (types.Value, error)) *FunctionValue {
	return &FunctionValue{name, env, f}
}

func (f *FunctionValue) Call(params []types.Value) (types.Value, error) {
	return f.function(params)
}

func (f *FunctionValue) String() string {
	return fmt.Sprintf("#<function %q>", f.name)
}
func (f *FunctionValue) Eval(_ types.Env) (types.Value, error) { return f, nil }
