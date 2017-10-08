package macroexpand

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinarvk/heisenlisp/lisperr"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/cons"
	"github.com/steinarvk/heisenlisp/value/null"
	"github.com/steinarvk/heisenlisp/value/symbol"
)

var (
	metricNontrivialMacroexpands = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "nontrivial_macroexpands",
			Help:      "Nontrivial calls to macroexpand",
		},
	)
)

func init() {
	prometheus.MustRegister(metricNontrivialMacroexpands)
}

func symbolIsSpecial(name string) bool {
	return name == "quote" || name == "quasiquote"
}

func macroexpandNonmacroCons(e types.Env, consval types.Value) (types.Value, error) {
	if null.IsNil(consval) {
		return null.Nil, nil
	}

	// (a b c d)
	// for good measure also (a b c . d)
	// we've determined that "a" is not a macro.

	car, cdr, ok := cons.Decompose(consval)
	if !ok {
		panic("not a cons")
	}

	if name, err := symbol.Name(car); err == nil && symbolIsSpecial(name) {
		// stop macroexpansion here.
		return consval, nil
	}

	// Note that we should not macroexpand cdr, as that
	// would be macroexpanding (b c d), which involves
	// looking up b.

	carExpanded, err := Macroexpand(e, car)
	if err != nil {
		return nil, err
	}

	cdrExpanded, err := macroexpandNonmacroCons(e, cdr)
	if err != nil {
		return nil, err
	}

	result := cons.New(carExpanded, cdrExpanded)

	cons.MarkAsMacroexpanded(result)

	return result, nil
}

func MacroexpandMultiple(e types.Env, vs []types.Value) ([]types.Value, error) {
	var rv []types.Value
	for _, v := range vs {
		ve, err := Macroexpand(e, v)
		if err != nil {
			return nil, err
		}
		rv = append(rv, ve)
	}
	return rv, nil
}

func Macroexpand(e types.Env, v types.Value) (types.Value, error) {
	if cons.IsMarkedAsMacroexpanded(v) {
		return v, nil
	}

	car, cdr, ok := cons.Decompose(v)
	if !ok {
		return v, nil
	}

	metricNontrivialMacroexpands.Inc()

	if !symbol.Is(car) {
		return macroexpandNonmacroCons(e, v)
	}

	functionOrMacro, err := car.Eval(e)
	if err != nil {
		_, ok := err.(lisperr.UnboundVariable)
		if ok {
			return macroexpandNonmacroCons(e, v)
		}
		return nil, err
	}

	macro, ok := functionOrMacro.(types.Macro)
	if !ok {
		return macroexpandNonmacroCons(e, v)
	}

	params, err := cons.ToProperList(cdr)
	if err != nil {
		return nil, err
	}

	result, err := macro.Expand(params)
	if err != nil {
		return nil, err
	}

	finalResult, err := Macroexpand(e, result)
	if err != nil {
		return nil, err
	}

	cons.MarkAsMacroexpanded(finalResult)
	return finalResult, nil
}
