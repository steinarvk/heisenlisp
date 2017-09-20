package expr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/steinarvk/heisenlisp/types"
)

type Bool bool

var (
	TrueValue  = Bool(true)
	FalseValue = Bool(false)
)

func (b Bool) AtomEquals(other types.Atom) bool {
	o, ok := other.(Bool)
	return ok && o == b
}

func (b Bool) Falsey() bool     { return !bool(b) }
func (_ Bool) TypeName() string { return "bool" }

func (b Bool) String() string {
	if bool(b) {
		return "true"
	}
	return "false"
}

func (b Bool) Eval(_ types.Env) (types.Value, error) { return b, nil }

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

func (_ Identifier) Falsey() bool { return false }

func (i Identifier) AtomEquals(other types.Atom) bool {
	o, ok := other.(Identifier)
	return ok && o == i
}

func (_ Identifier) TypeName() string { return "symbol" }

type Integer int64

func (i Integer) AtomEquals(other types.Atom) bool {
	o, ok := other.(Integer)
	return ok && o == i
}

func (i Integer) String() string {
	return fmt.Sprintf("%d", i)
}

func (i Integer) Eval(_ types.Env) (types.Value, error) { return i, nil }

func (i Integer) Falsey() bool { return i == 0 }

func (_ Integer) TypeName() string { return "integer" }

func (i Integer) AsInt64() (int64, bool)    { return int64(i), true }
func (i Integer) AsDouble() (float64, bool) { return float64(i), true }

type String string

func (s String) AtomEquals(other types.Atom) bool {
	o, ok := other.(String)
	return ok && o == s
}

func (s String) String() string {
	return fmt.Sprintf("%q", string(s))
}

func (s String) Eval(_ types.Env) (types.Value, error) { return s, nil }

func (s String) Falsey() bool     { return s == "" }
func (_ String) TypeName() string { return "string" }

type NilValue struct{}

func (_ NilValue) Falsey() bool                          { return true }
func (_ NilValue) String() string                        { return "nil" }
func (v NilValue) Eval(_ types.Env) (types.Value, error) { return v, nil }
func (v NilValue) AtomEquals(other types.Atom) bool {
	_, ok := other.(NilValue)
	return ok
}
func (_ NilValue) TypeName() string { return "nil" }

func IsNil(v types.Value) bool {
	_, ok := v.(NilValue)
	return ok
}

type ConsValue struct {
	car types.Value
	cdr types.Value
}

func Cons(a, b types.Value) *ConsValue {
	if a == nil {
		a = NilValue{}
	}
	if b == nil {
		b = NilValue{}
	}
	return &ConsValue{a, b}
}

func IsCons(v types.Value) bool {
	_, ok := v.(*ConsValue)
	return ok
}

func Car(v types.Value) (types.Value, error) {
	rv, ok := v.(*ConsValue)
	if !ok {
		return nil, errors.New("not a cons")
	}
	return rv.car, nil
}

func Cdr(v types.Value) (types.Value, error) {
	rv, ok := v.(*ConsValue)
	if !ok {
		return nil, errors.New("not a cons")
	}
	return rv.cdr, nil
}

func IsWrappedInUnary(name string, v types.Value) (types.Value, bool) {
	c, ok := v.(*ConsValue)
	if !ok {
		return nil, false
	}

	c2, ok := c.cdr.(*ConsValue)
	if !ok {
		return nil, false
	}
	if !IsNil(c2.cdr) {
		return nil, false
	}

	got, err := SymbolName(c.car)
	if err != nil {
		return nil, false
	}

	return c2.car, got == name
}

func ExpandQuasiQuote(e types.Env, mc types.Value) (types.Value, error) {
	c, ok := mc.(*ConsValue)
	if !ok {
		return mc, nil
	}

	head := Cons(nil, nil)
	tail := head

	if w, ok := IsWrappedInUnary("unquote", c.car); ok {
		newCar, err := w.Eval(e)
		if err != nil {
			return nil, err
		}
		head.car = newCar
	} else if w, ok := IsWrappedInUnary("unquote-splicing", c.car); ok {
		listToBeSpliced, err := w.Eval(e)
		if err != nil {
			return nil, err
		}

		elements, err := UnwrapList(listToBeSpliced)
		if err != nil {
			return nil, err
		}

		for i, elt := range elements {
			if i == 0 {
				head.car = elt
				tail = head
			} else {
				tail.cdr = Cons(elt, nil)
				tail = tail.cdr.(*ConsValue)
			}
		}
	} else {
		newCar, err := ExpandQuasiQuote(e, c.car)
		if err != nil {
			return nil, err
		}
		head.car = newCar
	}

	newCdr, err := ExpandQuasiQuote(e, c.cdr)
	if err != nil {
		return nil, err
	}
	tail.cdr = newCdr

	return head, nil
}

func (c *ConsValue) asProperList() ([]types.Value, bool) {
	var rv []types.Value

	node := c
	for {
		rv = append(rv, node.car)
		if IsNil(node.cdr) {
			return rv, true
		}

		next, ok := node.cdr.(*ConsValue)
		if !ok {
			return nil, false
		}
		node = next
	}
}

func (_ *ConsValue) TypeName() string { return "cons" }
func (c *ConsValue) Falsey() bool     { return false }
func (c *ConsValue) String() string {
	xs := []string{}

	node := c
	for {
		xs = append(xs, node.car.String())
		next, ok := node.cdr.(*ConsValue)
		if !ok {
			if !IsNil(node.cdr) {
				xs = append(xs, ".")
				xs = append(xs, node.cdr.String())
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
		if !specialForm.IsPure() && e.IsInPureContext() {
			return nil, errors.New("impure call in pure context")
		}
		return specialForm.Execute(e, unevaluatedParams)
	}

	macro, ok := funcVal.(types.Macro)
	if ok {
		if !macro.IsPure() && e.IsInPureContext() {
			return nil, errors.New("impure call in pure context")
		}

		newForm, err := macro.Expand(unevaluatedParams)
		if err != nil {
			return nil, err
		}
		return newForm.Eval(e)
	}

	callable, ok := funcVal.(types.Callable)
	if !ok {
		return nil, fmt.Errorf("%q (%v) is not callable", l[0], funcVal)
	}
	if !callable.IsPure() && e.IsInPureContext() {
		return nil, errors.New("impure call in pure context")
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
	pure     bool
}

func NewBuiltinFunction(name string, pure bool, f func([]types.Value) (types.Value, error)) *BuiltinFunctionValue {
	return &BuiltinFunctionValue{name, f, pure}
}

func (f *BuiltinFunctionValue) IsPure() bool { return f.pure }

func (_ *BuiltinFunctionValue) TypeName() string { return "function" }
func (f *BuiltinFunctionValue) Call(params []types.Value) (types.Value, error) {
	return f.function(params)
}

func (f *BuiltinFunctionValue) String() string {
	return fmt.Sprintf("#<builtin function %q>", f.name)
}
func (f *BuiltinFunctionValue) Eval(_ types.Env) (types.Value, error) { return f, nil }

func (f *BuiltinFunctionValue) Falsey() bool { return false }

func IntegerValue(v types.Value) (int64, error) {
	rv, ok := v.(Integer)
	if !ok {
		return 0, errors.New("not an integer")
	}
	return int64(rv), nil
}

func StringValue(v types.Value) (string, error) {
	rv, ok := v.(String)
	if !ok {
		return "", errors.New("not a string")
	}
	return string(rv), nil
}

func SymbolName(v types.Value) (string, error) {
	rv, ok := v.(Identifier)
	if !ok {
		return "", errors.New("not a symbol")
	}
	return string(rv), nil
}

func WrapList(vs []types.Value) types.Value {
	if len(vs) == 0 {
		return NilValue{}
	}
	var head, tail *ConsValue
	for _, v := range vs {
		nc := &ConsValue{v, NilValue{}}
		if head == nil {
			head = nc
			tail = nc
		} else {
			tail.cdr = nc
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

func UnwrapFixedList(v types.Value, l int) ([]types.Value, error) {
	xs, err := UnwrapList(v)
	if err != nil {
		return nil, err
	}
	if len(xs) != l {
		suffix := "s"
		if l == 1 {
			suffix = ""
		}
		return nil, fmt.Errorf("unwrapping list: expected %d element%s got %d", l, suffix, len(xs))
	}
	return xs, nil
}

func UnwrapProperListPair(v types.Value) (types.Value, types.Value, error) {
	xs, err := UnwrapFixedList(v, 2)
	if err != nil {
		return nil, nil, err
	}
	return xs[0], xs[1], nil
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

func ToSymbol(s string) types.Value {
	return Identifier(strings.ToLower(s))
}

func WrapInUnary(name string, v types.Value) types.Value {
	return Cons(ToSymbol(name), Cons(v, nil))
}

func AtomEquals(a, b types.Value) bool {
	av, aok := a.(types.Atom)
	bv, bok := b.(types.Atom)
	return aok && bok && av.AtomEquals(bv)
}

func ternaryAnd(a, b types.TernaryTruthValue) types.TernaryTruthValue {
	switch {
	case a == types.False || b == types.False:
		return types.False
	case a == types.True && b == types.True:
		return types.True
	default:
		return types.Maybe
	}
}

func Equals(a, b types.Value) (types.TernaryTruthValue, error) {
	if AtomEquals(a, b) {
		return types.True, nil
	}

	if IsCons(a) && IsCons(b) {
		ac := a.(*ConsValue)
		bc := b.(*ConsValue)
		tv1, err := Equals(ac.car, bc.car)
		if err != nil {
			return types.InvalidTernary, err
		}
		if tv1 == types.False {
			return types.False, nil
		}
		tv2, err := Equals(ac.cdr, bc.cdr)
		if err != nil {
			return types.InvalidTernary, err
		}
		return ternaryAnd(tv1, tv2), nil
	}

	unkA, okA := a.(types.Unknown)
	if okA {
		ok, err := unkA.Intersects(b)
		if err != nil {
			return types.InvalidTernary, err
		}
		if ok {
			return types.Maybe, nil
		}
		return types.False, nil
	}

	unkB, okB := b.(types.Unknown)
	if okB {
		ok, err := unkB.Intersects(a)
		if err != nil {
			return types.InvalidTernary, err
		}
		if ok {
			return types.Maybe, nil
		}
		return types.False, nil
	}

	return types.False, nil
}
