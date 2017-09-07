package expr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/steinarvk/heisenlisp/env"
	"github.com/steinarvk/heisenlisp/types"
)

type Bool bool

func (b Bool) AtomEquals(other types.Atom) bool {
	o, ok := other.(Bool)
	return ok && o == b
}

func (b Bool) Falsey() bool    { return !bool(b) }
func (b Bool) Uncertain() bool { return false }

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
func (a AnyOf) Uncertain() bool                       { return false }
func (a AnyOf) Falsey() bool                          { return false }

type FullyUnknown struct{}

func (_ FullyUnknown) String() string                        { return "#unknown" }
func (f FullyUnknown) Eval(_ types.Env) (types.Value, error) { return f, nil }
func (_ FullyUnknown) Falsey() bool                          { return false }
func (_ FullyUnknown) Uncertain() bool                       { return true }

// todo rename "symbol"
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

func (_ Identifier) Falsey() bool    { return false }
func (_ Identifier) Uncertain() bool { return false }

func (i Identifier) AtomEquals(other types.Atom) bool {
	o, ok := other.(Identifier)
	return ok && o == i
}

type Integer int64

func (i Integer) AtomEquals(other types.Atom) bool {
	o, ok := other.(Integer)
	return ok && o == i
}

func (i Integer) String() string {
	return fmt.Sprintf("%d", i)
}

func (i Integer) Eval(_ types.Env) (types.Value, error) { return i, nil }

func (i Integer) Falsey() bool    { return i == 0 }
func (i Integer) Uncertain() bool { return false }

type String string

func (s String) AtomEquals(other types.Atom) bool {
	o, ok := other.(String)
	return ok && o == s
}

func (s String) String() string {
	return fmt.Sprintf("%q", string(s))
}

func (s String) Eval(_ types.Env) (types.Value, error) { return s, nil }

func (s String) Falsey() bool    { return s == "" }
func (s String) Uncertain() bool { return false }

type NilValue struct{}

func (_ NilValue) Falsey() bool                          { return true }
func (_ NilValue) Uncertain() bool                       { return false }
func (_ NilValue) String() string                        { return "#nil" }
func (v NilValue) Eval(_ types.Env) (types.Value, error) { return v, nil }

func IsNil(v types.Value) bool {
	_, ok := v.(NilValue)
	return ok
}

type ConsValue struct {
	Car types.Value
	Cdr types.Value
}

func (c *ConsValue) asProperList() ([]types.Value, bool) {
	var rv []types.Value

	node := c
	for {
		rv = append(rv, node.Car)
		if IsNil(node.Cdr) {
			return rv, true
		}

		next, ok := node.Cdr.(*ConsValue)
		if !ok {
			return nil, false
		}
		node = next
	}
}

func (c *ConsValue) Falsey() bool { return false }
func (c *ConsValue) Uncertain() bool {
	return c.Car.Uncertain() || c.Cdr.Uncertain()
}
func (c *ConsValue) String() string {
	xs := []string{}

	node := c
	for {
		xs = append(xs, node.Car.String())
		next, ok := node.Cdr.(*ConsValue)
		if !ok {
			if !IsNil(node.Cdr) {
				xs = append(xs, ".")
				xs = append(xs, node.Cdr.String())
			}
			break
		}

		node = next
	}

	return fmt.Sprintf("(%s)", strings.Join(xs, " "))
}

func (c *ConsValue) Eval(e types.Env) (types.Value, error) {
	l, ok := c.asProperList()
	if !ok {
		return nil, fmt.Errorf("not a proper list")
	}

	if len(l) < 1 {
		return nil, errors.New("cannot evaluate empty list")
	}
	funcVal, err := l[0].Eval(e)
	if err != nil {
		return nil, err
	}

	unevaluatedParams := l[1:]

	specialForm, ok := funcVal.(types.SpecialForm)
	if ok {
		return specialForm.Execute(e, unevaluatedParams)
	}

	macro, ok := funcVal.(types.Macro)
	if ok {
		newForm, err := macro.Expand(e, unevaluatedParams)
		if err != nil {
			return nil, err
		}
		return newForm.Eval(e)
	}

	callable, ok := funcVal.(types.Callable)
	if !ok {
		return nil, fmt.Errorf("%q (%v) is not callable", l[0], funcVal)
	}

	var params []types.Value
	for _, unevaled := range unevaluatedParams {
		evaled, err := unevaled.Eval(e)
		if err != nil {
			return nil, err
		}
		params = append(params, evaled)
	}

	return callable.Call(params)
}

type BuiltinFunctionValue struct {
	name     string
	function func([]types.Value) (types.Value, error)
}

func NewBuiltinFunction(name string, f func([]types.Value) (types.Value, error)) *BuiltinFunctionValue {
	return &BuiltinFunctionValue{name, f}
}

func (f *BuiltinFunctionValue) Call(params []types.Value) (types.Value, error) {
	return f.function(params)
}

func (f *BuiltinFunctionValue) String() string {
	return fmt.Sprintf("#<builtin function %q>", f.name)
}
func (f *BuiltinFunctionValue) Eval(_ types.Env) (types.Value, error) { return f, nil }

func (f *BuiltinFunctionValue) Falsey() bool    { return false }
func (f *BuiltinFunctionValue) Uncertain() bool { return false }

type LispFunctionValue struct {
	name         string
	lexicalEnv   types.Env
	formalParams []string
	body         []types.Value
}

func NewLispFunction(env types.Env, name string, formalParams []string, body []types.Value) *LispFunctionValue {
	return &LispFunctionValue{name, env, formalParams, body}
}

func (f *LispFunctionValue) errorprefix() string {
	if f.name == "" {
		return "(anonymous function): "
	}
	return fmt.Sprintf("%s: ", f.name)
}

func (f *LispFunctionValue) Call(params []types.Value) (types.Value, error) {
	var rv types.Value
	var err error

	env := env.New(f.lexicalEnv)

	if len(params) != len(f.formalParams) {
		return nil, fmt.Errorf("%swant %d params, got %d", f.errorprefix(), len(f.formalParams), len(params))
	}
	for i, name := range f.formalParams {
		env.Bind(name, params[i])
	}

	for _, stmt := range f.body {
		rv, err = stmt.Eval(env)
		if err != nil {
			return nil, err
		}
	}
	return rv, nil
}

func (f *LispFunctionValue) String() string {
	if f.name == "" {
		return "#<anonymous function>"
	}
	return fmt.Sprintf("#<function %q>", f.name)
}
func (f *LispFunctionValue) Eval(_ types.Env) (types.Value, error) { return f, nil }

func (f *LispFunctionValue) Falsey() bool    { return false }
func (f *LispFunctionValue) Uncertain() bool { return false }

func SymbolName(v types.Value) (string, error) {
	rv, ok := v.(Identifier)
	if !ok {
		return "", errors.New("not a symbol")
	}
	return string(rv), nil
}

func WrapList(vs []types.Value) types.Value {
	var head, tail *ConsValue
	for _, v := range vs {
		nc := &ConsValue{v, NilValue{}}
		if head == nil {
			head = nc
			tail = nc
		} else {
			tail.Cdr = nc
			tail = nc
		}
	}
	return head
}

func UnwrapList(v types.Value) ([]types.Value, error) {
	if IsNil(v) {
		return nil, nil
	}

	rv, ok := v.(*ConsValue)
	if !ok {
		return nil, errors.New("not a list")
	}

	rrv, ok := rv.asProperList()
	if !ok {
		return nil, errors.New("not a list")
	}

	return rrv, nil
}

func Progn(e types.Env, vs []types.Value) (types.Value, error) {
	if len(vs) == 0 {
		return nil, errors.New("no body")
	}
	var result types.Value
	var err error
	for _, v := range vs {
		result, err = v.Eval(e)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
