package expr

import (
	"errors"
	"fmt"

	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/cons"
	"github.com/steinarvk/heisenlisp/value/integer"
	"github.com/steinarvk/heisenlisp/value/null"
	"github.com/steinarvk/heisenlisp/value/str"
	"github.com/steinarvk/heisenlisp/value/symbol"
)

func IsNil(v types.Value) bool {
	return null.IsNil(v)
}

func IsCons(v types.Value) bool {
	return cons.IsCons(v)
}

func WrapList(vs []types.Value) types.Value {
	return cons.FromProperList(vs)
}

func IsWrappedInUnary(name string, v types.Value) (types.Value, bool) {
	firstCar, firstCdr, ok := cons.Decompose(v)
	if !ok {
		return nil, false
	}

	secondCar, secondCdr, ok := cons.Decompose(firstCdr)

	if !IsNil(secondCdr) {
		return nil, false
	}

	got, err := symbol.Name(firstCar)
	if err != nil {
		return nil, false
	}

	return secondCar, got == name
}

func ExpandQuasiQuote(e types.Env, mc types.Value) (types.Value, error) {
	car, cdr, ok := cons.Decompose(mc)
	if !ok {
		return mc, nil
	}

	var carItems []types.Value

	if w, ok := IsWrappedInUnary("unquote", car); ok {
		newCar, err := w.Eval(e)
		if err != nil {
			return nil, err
		}
		carItems = append(carItems, newCar)
	} else if w, ok := IsWrappedInUnary("unquote-splicing", car); ok {
		listToBeSpliced, err := w.Eval(e)
		if err != nil {
			return nil, err
		}

		elements, err := UnwrapList(listToBeSpliced)
		if err != nil {
			return nil, err
		}

		for _, elt := range elements {
			carItems = append(carItems, elt)
		}
	} else {
		newCar, err := ExpandQuasiQuote(e, car)
		if err != nil {
			return nil, err
		}
		carItems = append(carItems, newCar)
	}

	newCdr, err := ExpandQuasiQuote(e, cdr)
	if err != nil {
		return nil, err
	}

	return cons.NewChain(carItems, newCdr), nil
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
	return integer.ToInt64(v)
}

func StringValue(v types.Value) (string, error) {
	return str.ToString(v)
}

func UnwrapList(v types.Value) ([]types.Value, error) {
	return cons.ToProperList(v)
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
	return symbol.New(s)
}

func WrapInUnary(name string, v types.Value) types.Value {
	return cons.New(ToSymbol(name), cons.New(v, nil))
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
		acar, acdr, _ := cons.Decompose(a)
		bcar, bcdr, _ := cons.Decompose(b)

		tv1, err := Equals(acar, bcar)
		if err != nil {
			return types.InvalidTernary, err
		}

		if tv1 == types.False {
			return types.False, nil
		}

		tv2, err := Equals(acdr, bcdr)
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
