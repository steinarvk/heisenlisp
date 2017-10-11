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
	parent types.Env

	singleBinding      uint32
	singleBindingValue types.Value

	bindings map[uint32]types.Value

	pureContext bool
}

func New(parent types.Env) types.Env {
	metricNewEnvironments.Inc()
	rv := &env{
		parent: parent,
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

func (e *env) Bind(k uint32, v types.Value) {
	metricEnvValueBinds.Inc()
	if e.bindings == nil {
		if e.singleBinding == 0 {
			e.singleBinding = k
			e.singleBindingValue = v
		} else {
			e.bindings = map[uint32]types.Value{
				e.singleBinding: e.singleBindingValue,
				k:               v,
			}
			e.singleBinding = 0
		}
		return
	}
	e.bindings[k] = v
}

func (e *env) BindRoot(k uint32, v types.Value) {
	if e.parent == nil {
		e.Bind(k, v)
		return
	}
	e.parent.BindRoot(k, v)
}

func (e *env) Lookup(k uint32) (types.Value, bool) {
	if e.singleBinding == k {
		return e.singleBindingValue, true
	}
	if e.bindings != nil {
		rv, ok := e.bindings[k]
		if ok {
			return rv, true
		}
	}
	if e.parent == nil {
		return nil, false
	}
	return e.parent.Lookup(k)
}
