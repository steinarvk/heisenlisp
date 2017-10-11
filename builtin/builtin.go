package builtin

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/env"
	"github.com/steinarvk/heisenlisp/equality"
	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/lisperr"
	"github.com/steinarvk/heisenlisp/listops"
	"github.com/steinarvk/heisenlisp/logic"
	"github.com/steinarvk/heisenlisp/numerics"
	"github.com/steinarvk/heisenlisp/purity"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/unknown"
	"github.com/steinarvk/heisenlisp/value/boolean"
	"github.com/steinarvk/heisenlisp/value/builtinfunc"
	"github.com/steinarvk/heisenlisp/value/cons"
	"github.com/steinarvk/heisenlisp/value/function"
	"github.com/steinarvk/heisenlisp/value/integer"
	"github.com/steinarvk/heisenlisp/value/macro"
	"github.com/steinarvk/heisenlisp/value/null"
	"github.com/steinarvk/heisenlisp/value/real"
	"github.com/steinarvk/heisenlisp/value/str"
	"github.com/steinarvk/heisenlisp/value/symbol"
	"github.com/steinarvk/heisenlisp/value/unknowns/anyof"
	"github.com/steinarvk/heisenlisp/value/unknowns/fullyunknown"
	"github.com/steinarvk/heisenlisp/value/unknowns/numinrange"
	"github.com/steinarvk/heisenlisp/value/unknowns/typed"

	_ "github.com/steinarvk/heisenlisp/cyclebreaker/impl"
)

var (
	Verbose = false
)

func wrap(name string, f func(a []types.Value) (types.Value, error)) func([]types.Value) (types.Value, error) {
	fw := func(vs []types.Value) (types.Value, error) {
		rv, err := f(vs)
		if err != nil {
			return nil, err
			// return nil, fmt.Errorf("%s: %v", name, err)
		}
		return rv, nil
	}
	return fw
}

func Unary(e types.Env, name string, f func(a types.Value) (types.Value, error)) {
	checker := func(vs []types.Value) (types.Value, error) {
		if len(vs) != 1 {
			return nil, fmt.Errorf("want 1 param, got %d", len(vs))
		}
		return f(vs[0])
	}
	e.Bind(symbol.StringToIdOrPanic(name), builtinfunc.New(name, purity.NameIsPure(name), wrap(name, checker)))
}

func Binary(e types.Env, name string, f func(a, b types.Value) (types.Value, error)) {
	checker := func(vs []types.Value) (types.Value, error) {
		if len(vs) != 2 {
			return nil, fmt.Errorf("want 2 params, got %d", len(vs))
		}
		return f(vs[0], vs[1])
	}
	e.Bind(symbol.StringToIdOrPanic(name), builtinfunc.New(name, purity.NameIsPure(name), wrap(name, checker)))
}

func Ternary(e types.Env, name string, f func(a, b, c types.Value) (types.Value, error)) {
	checker := func(vs []types.Value) (types.Value, error) {
		if len(vs) != 3 {
			return nil, fmt.Errorf("want 3 params, got %d", len(vs))
		}
		return f(vs[0], vs[1], vs[2])
	}
	e.Bind(symbol.StringToIdOrPanic(name), builtinfunc.New(name, purity.NameIsPure(name), wrap(name, checker)))
}

func Values(e types.Env, name string, f func(xs []types.Value) (types.Value, error)) {
	e.Bind(symbol.StringToIdOrPanic(name), builtinfunc.New(name, purity.NameIsPure(name), f))
}

func specialFormString(s string) string { return fmt.Sprintf("#<special %q>", s) }

type ifSpecialForm struct{}

func (i ifSpecialForm) IsPure() bool                        { return true }
func (i ifSpecialForm) Hashcode() uint32                    { return hashcode.Hash("special:", []byte(i.String())) }
func (i ifSpecialForm) TypeName() string                    { return "special" }
func (i ifSpecialForm) String() string                      { return specialFormString("if") }
func (i ifSpecialForm) Falsey() bool                        { return false }
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
	case types.Maybe:
		thenVal, err := thenClause.Eval(e)
		if err != nil {
			return nil, err
		}
		elseVal, err := elseClause.Eval(e)
		if err != nil {
			return nil, err
		}

		return anyof.New([]types.Value{
			thenVal, elseVal,
		})
	case types.True:
		return thenClause.Eval(e)
	case types.False:
		return elseClause.Eval(e)
	}
	return nil, errors.New("impossible state: ternary truth value neither true, false, or maybe")
}

type setSpecialForm struct{}

func (i setSpecialForm) TypeName() string                    { return "special" }
func (i setSpecialForm) IsPure() bool                        { return false }
func (i setSpecialForm) Hashcode() uint32                    { return hashcode.Hash("special:", []byte(i.String())) }
func (i setSpecialForm) String() string                      { return specialFormString("set!") }
func (i setSpecialForm) Falsey() bool                        { return false }
func (i setSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i setSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) != 2 {
		return nil, fmt.Errorf("'set!' expects 2 params, got %d", len(unevaluated))
	}

	name, err := symbol.Name(unevaluated[0])
	if err != nil {
		return nil, fmt.Errorf("must (set!) symbol, not %v", unevaluated[0])
	}

	value, err := unevaluated[1].Eval(e)
	if err != nil {
		return nil, err
	}

	e.BindRoot(symbol.StringToIdOrPanic(name), value)

	return value, nil
}

type quoteSpecialForm struct{}

func (i quoteSpecialForm) TypeName() string                    { return "special" }
func (i quoteSpecialForm) IsPure() bool                        { return true }
func (i quoteSpecialForm) Hashcode() uint32                    { return hashcode.Hash("special:", []byte(i.String())) }
func (i quoteSpecialForm) String() string                      { return specialFormString("quote") }
func (i quoteSpecialForm) Falsey() bool                        { return false }
func (i quoteSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i quoteSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) != 1 {
		return nil, fmt.Errorf("quote: unary special form, got %d params", len(unevaluated))
	}
	return unevaluated[0], nil
}

type quasiquoteSpecialForm struct{}

func (i quasiquoteSpecialForm) TypeName() string                    { return "special" }
func (i quasiquoteSpecialForm) IsPure() bool                        { return true }
func (i quasiquoteSpecialForm) Hashcode() uint32                    { return hashcode.Hash("special:", []byte(i.String())) }
func (i quasiquoteSpecialForm) String() string                      { return specialFormString("quasiquote") }
func (i quasiquoteSpecialForm) Falsey() bool                        { return false }
func (i quasiquoteSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i quasiquoteSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) != 1 {
		return nil, fmt.Errorf("quasiquote: unary special form, got %d params", len(unevaluated))
	}
	return expr.ExpandQuasiQuote(e, unevaluated[0])
}

type defunSpecialForm struct{}

func (i defunSpecialForm) TypeName() string                    { return "special" }
func (i defunSpecialForm) IsPure() bool                        { return false }
func (i defunSpecialForm) Hashcode() uint32                    { return hashcode.Hash("special:", []byte(i.String())) }
func (i defunSpecialForm) String() string                      { return specialFormString("defun!") }
func (i defunSpecialForm) Falsey() bool                        { return false }
func (i defunSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i defunSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) < 3 {
		return nil, fmt.Errorf("defun!: too few arguments")
	}

	name, err := symbol.Name(unevaluated[0])
	if err != nil {
		return nil, fmt.Errorf("must defun! symbol as name, not %v", unevaluated[0])
	}

	funcVal, err := makeFunction(&name, e, unevaluated[1], unevaluated[2:])
	if err != nil {
		return nil, err
	}

	e.BindRoot(symbol.StringToIdOrPanic(name), funcVal)
	return funcVal, nil
}

type defmacroSpecialForm struct{}

func (i defmacroSpecialForm) TypeName() string                    { return "special" }
func (i defmacroSpecialForm) IsPure() bool                        { return false }
func (i defmacroSpecialForm) Hashcode() uint32                    { return hashcode.Hash("special:", []byte(i.String())) }
func (i defmacroSpecialForm) String() string                      { return specialFormString("defmacro!") }
func (i defmacroSpecialForm) Falsey() bool                        { return false }
func (i defmacroSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i defmacroSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	if len(unevaluated) < 3 {
		return nil, fmt.Errorf("defmacro!: too few arguments")
	}

	name, err := symbol.Name(unevaluated[0])
	if err != nil {
		return nil, fmt.Errorf("defmacro! name: %v", err)
	}

	macroValue, err := macro.New(e, name, unevaluated[1], unevaluated[2:])
	if err != nil {
		return nil, err
	}

	e.BindRoot(symbol.StringToIdOrPanic(name), macroValue)

	return macroValue, nil
}

type letSpecialForm struct{}

func (i letSpecialForm) TypeName() string                    { return "special" }
func (i letSpecialForm) IsPure() bool                        { return true }
func (i letSpecialForm) Hashcode() uint32                    { return hashcode.Hash("special:", []byte(i.String())) }
func (i letSpecialForm) String() string                      { return specialFormString("let") }
func (i letSpecialForm) Falsey() bool                        { return false }
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

		sym, err := symbol.Id(bindingList[0])
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

type handleExceptionSpecialForm struct{}

func (i handleExceptionSpecialForm) TypeName() string { return "special" }
func (i handleExceptionSpecialForm) IsPure() bool     { return true }
func (i handleExceptionSpecialForm) Hashcode() uint32 {
	return hashcode.Hash("special:", []byte(i.String()))
}
func (i handleExceptionSpecialForm) String() string                      { return specialFormString("handle-exception") }
func (i handleExceptionSpecialForm) Falsey() bool                        { return false }
func (i handleExceptionSpecialForm) Eval(types.Env) (types.Value, error) { return i, nil }
func (i handleExceptionSpecialForm) Execute(e types.Env, unevaluated []types.Value) (types.Value, error) {
	// (handle-exception (progn body) ((x) (list "error:" x)))
	if len(unevaluated) != 2 {
		return nil, fmt.Errorf("handle-exception: not exactly two arguments")
	}

	errorHandlerDefinition, err := expr.UnwrapFixedList(unevaluated[1], 2)
	if err != nil {
		return nil, err
	}

	errorHandlerBody := errorHandlerDefinition[1]

	errorArgList, err := expr.UnwrapFixedList(errorHandlerDefinition[0], 1)
	if err != nil {
		return nil, err
	}

	errorVarName, err := symbol.Id(errorArgList[0])
	if err != nil {
		return nil, err
	}

	val, err := unevaluated[0].Eval(e)
	if err == nil {
		return val, nil
	}

	if exc, ok := err.(lisperr.LispException); ok {
		childEnv := env.New(e)
		childEnv.Bind(errorVarName, exc.Value())
		return errorHandlerBody.Eval(childEnv)
	}

	return nil, err
}

type andSpecialForm struct{}

func (i andSpecialForm) TypeName() string                    { return "special" }
func (i andSpecialForm) IsPure() bool                        { return true }
func (i andSpecialForm) Hashcode() uint32                    { return hashcode.Hash("special:", []byte(i.String())) }
func (i andSpecialForm) String() string                      { return specialFormString("and") }
func (i andSpecialForm) Falsey() bool                        { return false }
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
		case types.True:
			break
		case types.False:
			return boolean.False, nil
		default:
			knownToMaybeBeFalse = true
		}
	}

	if knownToMaybeBeFalse {
		return unknown.MaybeValue, nil
	}

	return boolean.True, nil
}

type orSpecialForm struct{}

func (i orSpecialForm) TypeName() string                    { return "special" }
func (i orSpecialForm) IsPure() bool                        { return true }
func (i orSpecialForm) Hashcode() uint32                    { return hashcode.Hash("special:", []byte(i.String())) }
func (i orSpecialForm) String() string                      { return specialFormString("or") }
func (i orSpecialForm) Falsey() bool                        { return false }
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
		case types.True:
			return boolean.True, nil
		case types.False:
			break
		default:
			knownToMaybeBeTrue = true
		}
	}

	if knownToMaybeBeTrue {
		return unknown.MaybeValue, nil
	}

	return boolean.False, nil
}

type lambdaSpecialForm struct{}

func (i lambdaSpecialForm) TypeName() string                    { return "special" }
func (i lambdaSpecialForm) IsPure() bool                        { return true }
func (i lambdaSpecialForm) Hashcode() uint32                    { return hashcode.Hash("special:", []byte(i.String())) }
func (i lambdaSpecialForm) String() string                      { return specialFormString("lambda") }
func (i lambdaSpecialForm) Falsey() bool                        { return false }
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
	bind := func(name string, v types.Value) {
		e.Bind(symbol.StringToIdOrPanic(name), v)
	}
	bind("if", &ifSpecialForm{})
	bind("set!", &setSpecialForm{})
	bind("defun!", &defunSpecialForm{})
	bind("defmacro!", &defmacroSpecialForm{})
	bind("lambda", &lambdaSpecialForm{})
	bind("quote", &quoteSpecialForm{})
	bind("quasiquote", &quasiquoteSpecialForm{})
	bind("let", &letSpecialForm{})
	bind("and", &andSpecialForm{})
	bind("or", &orSpecialForm{})
	bind("handle-exception", &handleExceptionSpecialForm{})

	bind("nil", null.Nil)
	bind("true", boolean.True)
	bind("false", boolean.False)
	bind("maybe", unknown.MaybeValue)
	bind("unknown", fullyunknown.Value)

	bind("nan", real.FromFloat64(math.NaN()))

	Unary(e, "_nan?", func(a types.Value) (types.Value, error) {
		return boolean.FromBool(real.IsNaN(a)), nil
	})

	Unary(e, "_atom?", func(a types.Value) (types.Value, error) {
		_, ok := a.(types.Atom)
		return boolean.FromBool(ok), nil
	})

	Binary(e, "_atom-eq?", func(a, b types.Value) (types.Value, error) {
		return boolean.FromBool(equality.AtomEquals(a, b)), nil
	})

	Unary(e, "_type", func(a types.Value) (types.Value, error) {
		return expr.ToSymbol(a.TypeName()), nil
	})

	Unary(e, "_to-string", func(a types.Value) (types.Value, error) {
		return str.New(a.String()), nil
	})

	Binary(e, "_assert!", func(sval, anyval types.Value) (types.Value, error) {
		s, err := expr.StringValue(sval)
		if err != nil {
			return nil, err
		}

		srep := anyval.String()

		success := s == srep
		if success {
			if Verbose {
				log.Printf("PASS: %q == %q", s, srep)
			}
			return boolean.True, nil
		}

		if Verbose {
			log.Printf("FAIL: %q != %q", s, srep)
		}
		return nil, fmt.Errorf("assertion failed: %q != %q", s, srep)
	})

	Unary(e, "_uncertain?", func(a types.Value) (types.Value, error) {
		return boolean.FromBool(unknown.IsUncertain(a)), nil
	})

	checkEquality := func(a, b types.Value) (types.Value, error) {
		// TODO: non-binary equality checking in heisenlisp.
		// Consider: (= (any-of 0 1) (any-of 1 2) (any-of 0 3))
		// We need to precisely define what exactly this means.
		// I guess it _should_ mean taking the intersection of all of these?
		tv, err := equality.Equals(a, b)
		if err != nil {
			return nil, err
		}
		switch tv {
		case types.False:
			return boolean.False, nil
		case types.True:
			return boolean.True, nil
		case types.Maybe:
			return unknown.MaybeValue, nil
		}
		panic("impossible")
	}

	Binary(e, "equals?", checkEquality)
	Binary(e, "=", checkEquality)

	Binary(e, "cons", func(a, b types.Value) (types.Value, error) {
		return cons.New(a, b), nil
	})

	Unary(e, "car", func(a types.Value) (types.Value, error) {
		return cons.Car(a)
	})

	Unary(e, "cdr", func(a types.Value) (types.Value, error) {
		return cons.Cdr(a)
	})

	Unary(e, "not", func(a types.Value) (types.Value, error) {
		val, err := unknown.TruthValue(a)
		if err != nil {
			return nil, err
		}
		switch val {
		case types.False:
			return boolean.True, nil
		case types.True:
			return boolean.False, nil
		case types.Maybe:
			return unknown.MaybeValue, nil
		}
		return nil, errors.New("impossible state: ternary truth value neither true, false, or maybe")
	})

	Unary(e, "may?", func(a types.Value) (types.Value, error) {
		val, err := unknown.TruthValue(a)
		if err != nil {
			return nil, err
		}
		switch val {
		case types.False:
			return boolean.False, nil
		default:
			return boolean.True, nil
		}
	})

	Unary(e, "must?", func(a types.Value) (types.Value, error) {
		val, err := unknown.TruthValue(a)
		if err != nil {
			return nil, err
		}
		switch val {
		case types.True:
			return boolean.True, nil
		default:
			return boolean.False, nil
		}
	})

	Unary(e, "length", func(consish types.Value) (types.Value, error) {
		zero := integer.FromInt64(0)
		one := integer.FromInt64(1)
		f := func(lastCount, _ types.Value) (types.Value, error) {
			rv, err := numerics.BinaryPlus(lastCount, one)
			return rv, err
		}
		g := func(lastCount, tail types.Value) (types.Value, error) {
			_, ok := tail.(types.Unknown)
			if !ok {
				return nil, lisperr.UnexpectedValue{"valid list-tail", consish}
			}
			// TODO: do any sort of filtering of unknowns.
			return numinrange.New(lastCount.(types.Numeric), nil, true, false, []string{integer.TypeName})
		}
		return listops.FoldLeftWithOddTail(f, g, zero, consish)
	})

	Binary(e, "apply", func(f, args types.Value) (types.Value, error) {
		callable, ok := f.(types.Callable)
		if !ok {
			return nil, errors.New("not a callable")
		}

		xs, err := expr.UnwrapList(args)
		if err != nil {
			return nil, err
		}

		return callable.Call(xs)
	})

	Unary(e, "reversed", func(l types.Value) (types.Value, error) {
		return listops.Reversed(l)
	})

	Binary(e, "map", func(f, l types.Value) (types.Value, error) {
		callable, ok := f.(types.Callable)
		if !ok {
			return nil, lisperr.UnexpectedValue{"callable", f}
		}

		cb := func(a types.Value) (types.Value, error) {
			return callable.Call([]types.Value{a})
		}

		return listops.Map(cb, l)
	})

	Binary(e, "any?", func(f, l types.Value) (types.Value, error) {
		callable, ok := f.(types.Callable)
		if !ok {
			return nil, errors.New("not a callable")
		}

		cb := func(a, b types.Value) (types.Value, bool, error) {
			rv, err := callable.Call([]types.Value{b})
			if err != nil {
				return nil, false, err
			}
			either, err := logic.TernaryOrValues(a, rv)
			if err != nil {
				return nil, false, err
			}
			switch either {
			case types.False:
				return boolean.False, false, nil
			case types.True:
				return boolean.True, true, nil
			case types.Maybe:
				return anyof.MaybeValue, false, nil
			}
			panic("impossible")
		}

		return listops.FoldLeftShortcircuit(cb, boolean.False, l)
	})

	Binary(e, "all?", func(f, l types.Value) (types.Value, error) {
		callable, ok := f.(types.Callable)
		if !ok {
			return nil, errors.New("not a callable")
		}

		cb := func(a, b types.Value) (types.Value, bool, error) {
			rv, err := callable.Call([]types.Value{b})
			if err != nil {
				return nil, false, err
			}
			both, err := logic.TernaryAndValues(a, rv)
			if err != nil {
				return nil, false, err
			}
			switch both {
			case types.False:
				return boolean.False, true, nil
			case types.True:
				return boolean.True, false, nil
			case types.Maybe:
				return anyof.MaybeValue, false, nil
			}
			panic("impossible")
		}

		return listops.FoldLeftShortcircuit(cb, boolean.True, l)
	})

	Ternary(e, "reduce-left", func(f, initial, l types.Value) (types.Value, error) {
		xs, err := expr.UnwrapList(l)
		if err != nil {
			return nil, err
		}

		if len(xs) == 0 {
			return initial, nil
		}

		if len(xs) == 1 {
			return xs[0], nil
		}

		callable, ok := f.(types.Callable)
		if !ok {
			return nil, errors.New("not a callable")
		}

		reduced, err := callable.Call([]types.Value{xs[0], xs[1]})
		if err != nil {
			return nil, err
		}

		for _, x := range xs[2:] {
			next, err := callable.Call([]types.Value{reduced, x})
			if err != nil {
				return nil, err
			}
			reduced = next
		}

		return reduced, nil
	})

	Unary(e, "list?", func(l types.Value) (types.Value, error) {
		return listops.Foldable(l), nil
	})

	Binary(e, "filter", func(f, l types.Value) (types.Value, error) {
		callable, ok := f.(types.Callable)
		if !ok {
			return nil, lisperr.UnexpectedValue{"callable", f}
		}

		cb := func(a types.Value) (types.Value, error) {
			return callable.Call([]types.Value{a})
		}

		return listops.Filter(cb, l)
	})

	Binary(e, "filter-reversed", func(f, l types.Value) (types.Value, error) {
		callable, ok := f.(types.Callable)
		if !ok {
			return nil, lisperr.UnexpectedValue{"callable", f}
		}

		cb := func(a types.Value) (types.Value, error) {
			return callable.Call([]types.Value{a})
		}

		return listops.FilterReversed(cb, l)
	})

	Ternary(e, "fold-right", func(f, initial, l types.Value) (types.Value, error) {
		// folding is like reduction, except the initial value is always used.

		callable, ok := f.(types.Callable)
		if !ok {
			return nil, lisperr.UnexpectedValue{"callable", f}
		}

		cb := func(a, b types.Value) (types.Value, error) {
			return callable.Call([]types.Value{a, b})
		}

		return listops.FoldRight(cb, initial, l)
	})

	Ternary(e, "fold-left", func(f, initial, l types.Value) (types.Value, error) {
		// folding is like reduction, except the initial value is always used.

		callable, ok := f.(types.Callable)
		if !ok {
			return nil, lisperr.UnexpectedValue{"callable", f}
		}

		cb := func(a, b types.Value) (types.Value, error) {
			return callable.Call([]types.Value{a, b})
		}

		return listops.FoldLeft(cb, initial, l)
	})

	Values(e, "trace", func(xs []types.Value) (types.Value, error) {
		// print function for debugging; really not a pure func
		var printsections []string
		var rv types.Value
		for _, x := range xs {
			printsections = append(printsections, x.String())
			rv = x
		}
		log.Printf("trace: %s", strings.Join(printsections, " "))
		return rv, nil
	})

	Values(e, "unknown-of-type", func(xs []types.Value) (types.Value, error) {
		var typenames []string

		for _, x := range xs {
			name, err := symbol.Name(x)
			if err != nil {
				return nil, err
			}
			typenames = append(typenames, name)
		}

		return typed.New(typenames...), nil
	})

	parseKeywords := func(xs []types.Value) (map[string]types.Value, error) {
		if len(xs)%2 != 0 {
			return nil, fmt.Errorf("parsing keywords from odd-numbered arguments: %v", xs)
		}

		rv := map[string]types.Value{}
		for i := 0; i < len(xs); i += 2 {
			name, err := symbol.Name(xs[i])
			if err != nil {
				return nil, fmt.Errorf("parsing keywords: %v", err)
			}

			rv[name] = xs[i+1]
		}

		return rv, nil
	}

	Values(e, "number-in-range", func(xs []types.Value) (types.Value, error) {
		args, err := parseKeywords(xs)
		if err != nil {
			return nil, err
		}

		// TODO check for unused arguments

		var lowVal, highVal types.Value
		var lowIncl, highIncl bool

		val, hasFrom := args["from"]
		if hasFrom {
			lowVal = val
			lowIncl = true
		}
		val, hasAbove := args["above"]
		if hasAbove {
			lowVal = val
			lowIncl = false
		}

		val, hasTo := args["to"]
		if hasTo {
			highVal = val
			highIncl = true
		}
		val, hasBelow := args["below"]
		if hasBelow {
			highVal = val
			highIncl = false
		}

		if hasFrom && hasAbove {
			return nil, errors.New("two lower bounds specified (from and above)")
		}
		if hasTo && hasBelow {
			return nil, errors.New("two upper bounds specified (to and below)")
		}

		low, ok := lowVal.(types.Numeric)
		if !ok && lowVal != nil {
			return nil, fmt.Errorf("invalid numeric bound: %v", lowVal)
		}

		high, ok := highVal.(types.Numeric)
		if !ok && highVal != nil {
			return nil, fmt.Errorf("invalid numeric bound: %v", highVal)
		}

		// TODO types
		return numinrange.New(low, high, lowIncl, highIncl, nil)
	})

	Values(e, "any-of", func(xs []types.Value) (types.Value, error) {
		return anyof.New(xs)
	})

	Unary(e, "type", func(v types.Value) (types.Value, error) {
		unk, ok := v.(types.Unknown)
		if !ok {
			return symbol.New(v.TypeName()), nil
		}

		typenames, ok := unk.ActualTypeName()
		if !ok {
			return typed.New(symbol.TypeName), nil
		}

		if len(typenames) == 0 {
			return nil, errors.New("no types returned")
		}

		if len(typenames) == 1 {
			return symbol.New(typenames[0]), nil
		}

		var typeSyms []types.Value
		for _, typename := range typenames {
			typeSyms = append(typeSyms, symbol.New(typename))
		}
		return anyof.New(typeSyms)
	})

	Unary(e, "possible-values", func(v types.Value) (types.Value, error) {
		vals, ok := anyof.PossibleValues(v)
		if ok {
			return expr.WrapList(vals), nil
		}
		return fullyunknown.Value, nil
	})

	Unary(e, "range", func(v types.Value) (types.Value, error) {
		n, err := integer.ToInt64(v)
		if err != nil {
			return nil, err
		}

		var xs []types.Value
		for i := int64(0); i < n; i++ {
			xs = append(xs, integer.FromInt64(i))
		}
		return expr.WrapList(xs), nil
	})

	// note: this is a loophole from the side effect rule.
	Unary(e, "throw-exception", func(v types.Value) (types.Value, error) {
		return nil, lisperr.NewException(v)
	})

	Binary(e, "low-level-plus", numerics.BinaryPlus)
	Binary(e, "low-level-minus", numerics.BinaryMinus)
	Binary(e, "low-level-multiply", numerics.BinaryMultiply)
	Binary(e, "low-level-divide", numerics.BinaryDivision)
	Binary(e, "mod", numerics.Mod)

	Binary(e, "<=", numerics.BinaryLeq)
	Binary(e, "<", numerics.BinaryLess)
	Binary(e, ">", numerics.BinaryGreater)
	Binary(e, ">=", numerics.BinaryGeq)
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

func loadFile(e types.Env, fn string) error {
	_, err := code.RunFile(e, fn)
	return err
}

func loadStandardLibrary(e types.Env) error {
	// TODO: for compiled binaries, embed standard library

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	wd, err = filepath.Abs(wd)
	if err != nil {
		return err
	}

	for {
		_, err := os.Stat(filepath.Join(wd, "core"))
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return err
		}
		nwd := filepath.Dir(wd)
		if nwd == wd {
			return errors.New("standard library ./core not found")
		}
		wd = nwd
	}

	dir := filepath.Join(wd, "core")

	if Verbose {
		log.Printf("standard library directory: %q", dir)
	}

	fns, err := listLispFilesInOrder(dir)
	if err != nil {
		return err
	}
	if len(fns) == 0 {
		return errors.New("no standard library .hlisp files found")
	}

	if Verbose {
		log.Printf("standard library: %v", fns)
	}

	for _, fn := range fns {
		if Verbose {
			log.Printf("loading %q", fn)
		}
		if err := loadFile(e, fn); err != nil {
			return fmt.Errorf("error loading %q: %v", fn, err)
		}
	}
	return nil
}

func NewRootEnv() types.Env {
	rv := env.New(nil)
	BindDefaults(rv)
	if err := loadStandardLibrary(rv); err != nil {
		panic(fmt.Errorf("error loading standard library: %v", err))
	}
	return rv
}
