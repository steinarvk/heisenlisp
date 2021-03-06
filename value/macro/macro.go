package macro

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/lambdalist"
	"github.com/steinarvk/heisenlisp/macroexpand"
	"github.com/steinarvk/heisenlisp/purity"
	"github.com/steinarvk/heisenlisp/types"
)

const TypeName = "macro"

var (
	metricNewLispMacro = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "new_lisp_macro",
			Help:      "New lisp macro values created",
		},
	)

	metricLispMacroExpansions = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "lisp_macro_expansions",
			Help:      "Expansions of Lisp macros",
		},
	)
)

func init() {
	prometheus.MustRegister(metricNewLispMacro)
	prometheus.MustRegister(metricLispMacroExpansions)
}

type macroValue struct {
	name       string
	lexicalEnv types.Env
	lambdaList *lambdalist.LambdaList
	body       []types.Value
	pure       bool
}

func (f *macroValue) Expand(params []types.Value) (types.Value, error) {
	var rv types.Value
	var err error

	env, err := f.lambdaList.BindArgs(f.lexicalEnv, params, f.pure)
	if err != nil {
		return nil, fmt.Errorf("%s%v", f.errorprefix(), err)
	}

	for _, stmt := range f.body {
		rv, err = stmt.Eval(env)
		if err != nil {
			return nil, err
		}
	}

	metricLispMacroExpansions.Inc()
	return rv, nil
}

func (f *macroValue) IsPure() bool {
	return f.pure
}

func (f *macroValue) String() string {
	if f.name == "" {
		return "#<anonymous macro>"
	}
	return fmt.Sprintf("#<macro %q>", f.name)
}
func (f *macroValue) Eval(_ types.Env) (types.Value, error) { return f, nil }

func (f *macroValue) Falsey() bool { return false }

func New(env types.Env, name string, formalParams types.Value, body []types.Value) (*macroValue, error) {
	ll, err := lambdalist.Parse(formalParams)
	if err != nil {
		return nil, fmt.Errorf("invalid lambda list: %v", err)
	}

	expandedBody, err := macroexpand.MacroexpandMultiple(env, body)
	if err != nil {
		return nil, fmt.Errorf("error macroexpanding macro body: %v", err)
	}

	rv := &macroValue{
		name:       name,
		lexicalEnv: env,
		lambdaList: ll,
		body:       expandedBody,
	}
	if purity.NameIsPure(name) {
		rv.pure = true
	}
	metricNewLispMacro.Inc()
	return rv, nil
}

func (_ *macroValue) TypeName() string { return TypeName }
func (f *macroValue) errorprefix() string {
	if f.name == "" {
		return "(anonymous macro): "
	}
	return fmt.Sprintf("%s: ", f.name)
}

func (f *macroValue) Hashcode() uint32 {
	return hashcode.Hash(fmt.Sprintf("%p", f))
}
