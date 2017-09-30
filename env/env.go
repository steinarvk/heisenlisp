package env

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinarvk/heisenlisp/types"
)

var (
	metricNewEnvironments = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "new_environments",
			Help:      "New environments created",
		},
	)

	metricEnvValueBinds = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "new_env_value_bindings",
			Help:      "Values bound in environments",
		},
	)
)

func init() {
	prometheus.MustRegister(metricNewEnvironments)
	prometheus.MustRegister(metricEnvValueBinds)
}

type env struct {
	parent      types.Env
	bindings    map[string]types.Value
	pureContext bool
}

func New(parent types.Env) types.Env {
	metricNewEnvironments.Inc()
	rv := &env{
		parent:   parent,
		bindings: map[string]types.Value{},
	}
	if parent != nil && parent.IsInPureContext() {
		rv.pureContext = true
	}
	return rv
}

func (e *env) MarkPure() {
	e.pureContext = true
}

func (e *env) IsInPureContext() bool {
	return e.pureContext
}

func (e *env) Bind(k string, v types.Value) {
	metricEnvValueBinds.Inc()
	e.bindings[k] = v
}

func (e *env) BindRoot(k string, v types.Value) {
	if e.parent == nil {
		e.Bind(k, v)
		return
	}
	e.parent.BindRoot(k, v)
}

func (e *env) Lookup(k string) (types.Value, bool) {
	rv, ok := e.bindings[k]
	if ok {
		return rv, true
	}
	if e.parent == nil {
		return nil, false
	}
	return e.parent.Lookup(k)
}
