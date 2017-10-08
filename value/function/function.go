package function

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/lambdalist"
	"github.com/steinarvk/heisenlisp/macroexpand"
	"github.com/steinarvk/heisenlisp/purity"
	"github.com/steinarvk/heisenlisp/types"
)

var (
	metricNewLispFunction = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "new_lisp_function",
			Help:      "New lisp function values created",
		},
	)

	metricLispFunctionCall = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "lisp_function_calls",
			Help:      "Calls to Lisp functions",
		},
	)
)

func init() {
	prometheus.MustRegister(metricNewLispFunction)
	prometheus.MustRegister(metricLispFunctionCall)
}

type functionValue struct {
	name       string
	lexicalEnv types.Env
	lambdaList *lambdalist.LambdaList
	body       []types.Value
	pure       bool
}

func New(env types.Env, name string, formalParams types.Value, body []types.Value) (types.Value, error) {
	ll, err := lambdalist.Parse(formalParams)
	if err != nil {
		return nil, fmt.Errorf("invalid lambda list: %v", err)
	}

	expandedBody, err := macroexpand.MacroexpandMultiple(env, body)
	if err != nil {
		return nil, fmt.Errorf("error macroexpanding function body: %v", err)
	}

	rv := &functionValue{
		name:       name,
		lexicalEnv: env,
		lambdaList: ll,
		body:       expandedBody,
	}
	if purity.NameIsPure(name) {
		rv.pure = true
	}
	metricNewLispFunction.Inc()
	return rv, nil
}

func (_ *functionValue) TypeName() string { return "function" }
func (f *functionValue) errorprefix() string {
	if f.name == "" {
		return "(anonymous function): "
	}
	return fmt.Sprintf("%s: ", f.name)
}

func (f *functionValue) Call(params []types.Value) (types.Value, error) {
	var rv types.Value
	var err error

	env, err := f.lambdaList.BindArgs(f.lexicalEnv, params, f.pure)
	if err != nil {
		return nil, err
		// return nil, fmt.Errorf("%s%v", f.errorprefix(), err)
	}

	for _, stmt := range f.body {
		rv, err = stmt.Eval(env)
		if err != nil {
			return nil, err
		}
	}

	metricLispFunctionCall.Inc()
	return rv, nil
}

func (f *functionValue) String() string {
	if f.name == "" {
		return "#<anonymous function>"
	}
	return fmt.Sprintf("#<function %q>", f.name)
}
func (f *functionValue) Eval(_ types.Env) (types.Value, error) { return f, nil }

func (f *functionValue) IsPure() bool {
	return f.pure
}

func (f *functionValue) Falsey() bool { return false }

func (f *functionValue) Hashcode() uint32 {
	return hashcode.Hash(fmt.Sprintf("%p", f))
}
