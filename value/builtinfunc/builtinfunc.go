package builtinfunc

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/types"
)

const (
	TypeName = "function"
)

var (
	metricNewBuiltinFunction = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "new_builtin_function",
			Help:      "New builtin function values created",
		},
	)

	metricBuiltinFunctionCall = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "builtin_function_calls",
			Help:      "Calls to builtin functions",
		},
	)
)

func init() {
	prometheus.MustRegister(metricNewBuiltinFunction)
	prometheus.MustRegister(metricBuiltinFunctionCall)
}

type builtinFunctionValue struct {
	name     string
	function func([]types.Value) (types.Value, error)
	pure     bool
}

var _ types.Value = &builtinFunctionValue{}
var _ types.Callable = &builtinFunctionValue{}

func (f *builtinFunctionValue) CallableName() string { return f.name }

func New(name string, pure bool, f func([]types.Value) (types.Value, error)) types.Value {
	metricNewBuiltinFunction.Inc()
	return &builtinFunctionValue{name, f, pure}
}

func (f *builtinFunctionValue) IsPure() bool { return f.pure }

func (_ *builtinFunctionValue) TypeName() string { return TypeName }
func (f *builtinFunctionValue) Call(params []types.Value) (types.Value, error) {
	metricBuiltinFunctionCall.Inc()
	return f.function(params)
}

func (f *builtinFunctionValue) String() string {
	return fmt.Sprintf("#<builtin function %q>", f.name)
}
func (f *builtinFunctionValue) Eval(_ types.Env) (types.Value, error) { return f, nil }

func (f *builtinFunctionValue) Falsey() bool { return false }

func (f *builtinFunctionValue) Hashcode() uint32 {
	return hashcode.Hash(fmt.Sprintf("%p", f))
}
