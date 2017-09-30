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
