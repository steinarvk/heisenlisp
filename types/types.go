package types

type Value interface {
	String() string
	Eval(e Env) (Value, error)
}

type Callable interface {
	Value
	Call([]Value) (Value, error)
}

type Env interface {
	Bind(k string, v Value)
	Lookup(k string) (Value, bool)
}
