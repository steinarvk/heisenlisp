package types

type Value interface {
	String() string
	Falsey() bool
	Uncertain() bool
	Eval(e Env) (Value, error)
}

type SpecialForm interface {
	Value
	Execute(Env, []Value) (Value, error)
}

type Macro interface {
	Value
	Expand(Env, []Value) (Value, error)
}

type Callable interface {
	Value
	Call([]Value) (Value, error)
}

type Env interface {
	Bind(k string, v Value)
	Lookup(k string) (Value, bool)
}
