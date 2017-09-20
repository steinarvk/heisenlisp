package env

import "github.com/steinarvk/heisenlisp/types"

type env struct {
	parent      types.Env
	bindings    map[string]types.Value
	pureContext bool
}

func New(parent types.Env) types.Env {
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
	e.bindings[k] = v
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
