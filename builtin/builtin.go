package builtin

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/env"
	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/function"
	"github.com/steinarvk/heisenlisp/numerics"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/unknown"
)

var (
	Verbose = false
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

func Ternary(e types.Env, name string, f func(a, b, c types.Value) (types.Value, error)) {
	checker := func(vs []types.Value) (types.Value, error) {
		if len(vs) != 3 {
			return nil, fmt.Errorf("want 3 params, got %d", len(vs))
		}
		return f(vs[0], vs[1], vs[2])
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
	case types.Maybe:
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
		case types.True:
			break
		case types.False:
			return expr.FalseValue, nil
		default:
			knownToMaybeBeFalse = true
		}
	}

	if knownToMaybeBeFalse {
		return unknown.MaybeValue, nil
	}

	return expr.TrueValue, nil
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
		case types.True:
			return expr.TrueValue, nil
		case types.False:
			break
		default:
			knownToMaybeBeTrue = true
		}
	}

	if knownToMaybeBeTrue {
		return unknown.MaybeValue, nil
	}

	return expr.FalseValue, nil
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
	e.Bind("and", &andSpecialForm{})
	e.Bind("or", &orSpecialForm{})

	e.Bind("nil", expr.NilValue{})
	e.Bind("true", expr.TrueValue)
	e.Bind("false", expr.FalseValue)
	e.Bind("maybe", unknown.MaybeValue)
	e.Bind("unknown", unknown.FullyUnknown{})

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

	Unary(e, "_to-string", func(a types.Value) (types.Value, error) {
		return expr.String(a.String()), nil
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
			return expr.TrueValue, nil
		}

		if Verbose {
			log.Printf("FAIL: %q != %q", s, srep)
		}
		return nil, fmt.Errorf("assertion failed: %q != %q", s, srep)
	})

	Unary(e, "_uncertain?", func(a types.Value) (types.Value, error) {
		return expr.Bool(unknown.IsUncertain(a)), nil
	})

	checkEquality := func(a, b types.Value) (types.Value, error) {
		// TODO: non-binary equality checking in heisenlisp.
		// Consider: (= (any-of 0 1) (any-of 1 2) (any-of 0 3))
		// We need to precisely define what exactly this means.
		// I guess it _should_ mean taking the intersection of all of these?
		tv, err := expr.Equals(a, b)
		if err != nil {
			return nil, err
		}
		switch tv {
		case types.False:
			return expr.FalseValue, nil
		case types.True:
			return expr.TrueValue, nil
		case types.Maybe:
			return unknown.MaybeValue, nil
		}
		panic("impossible")
	}

	Binary(e, "equals?", checkEquality)
	Binary(e, "=", checkEquality)

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
		case types.False:
			return expr.TrueValue, nil
		case types.True:
			return expr.FalseValue, nil
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
			return expr.FalseValue, nil
		default:
			return expr.TrueValue, nil
		}
	})

	Unary(e, "must?", func(a types.Value) (types.Value, error) {
		val, err := unknown.TruthValue(a)
		if err != nil {
			return nil, err
		}
		switch val {
		case types.True:
			return expr.TrueValue, nil
		default:
			return expr.FalseValue, nil
		}
	})

	Unary(e, "length", func(a types.Value) (types.Value, error) {
		xs, err := expr.UnwrapList(a)
		if err != nil {
			return nil, err
		}
		return expr.Integer(len(xs)), nil
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

	Binary(e, "map", func(f, l types.Value) (types.Value, error) {
		// TODO: handle uncertainty
		callable, ok := f.(types.Callable)
		if !ok {
			return nil, errors.New("not a callable")
		}

		xs, err := expr.UnwrapList(l)
		if err != nil {
			return nil, err
		}

		var rv []types.Value
		for _, x := range xs {
			xm, err := callable.Call([]types.Value{x})
			if err != nil {
				return nil, err
			}
			rv = append(rv, xm)
		}

		return expr.WrapList(rv), nil
	})

	Ternary(e, "reduce-left", func(f, initial, l types.Value) (types.Value, error) {
		callable, ok := f.(types.Callable)
		if !ok {
			return nil, errors.New("not a callable")
		}

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

	Values(e, "any-of", func(xs []types.Value) (types.Value, error) {
		return unknown.NewMaybeAnyOf(xs)
	})

	Unary(e, "possible-values", func(v types.Value) (types.Value, error) {
		vals, ok := unknown.PossibleValues(v)
		if ok {
			return expr.WrapList(vals), nil
		}
		return unknown.FullyUnknown{}, nil
	})

	Unary(e, "range", func(v types.Value) (types.Value, error) {
		n, err := expr.IntegerValue(v)
		if err != nil {
			return nil, err
		}

		var xs []types.Value
		for i := int64(0); i < n; i++ {
			xs = append(xs, expr.Integer(i))
		}
		return expr.WrapList(xs), nil
	})

	Binary(e, "low-level-plus", numerics.BinaryPlus)
	Binary(e, "low-level-minus", numerics.BinaryMinus)
	Binary(e, "low-level-multiply", numerics.BinaryMultiply)
	Binary(e, "mod", numerics.Mod)
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

	fns, err := listLispFilesInOrder("./core")
	if err != nil {
		return err
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
