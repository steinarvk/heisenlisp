package types

type TernaryTruthValue string

const (
	False          = TernaryTruthValue("false")
	True           = TernaryTruthValue("true")
	Maybe          = TernaryTruthValue("maybe")
	InvalidTernary = TernaryTruthValue("invalid")
)

type Value interface {
	String() string
	Falsey() bool
	TypeName() string
	Eval(e Env) (Value, error)
}

type Unknown interface {
	ActualTypeName() ([]string, bool)
	Intersects(Value) (bool, error)
}

type Atom interface {
	Value
	AtomEquals(other Atom) bool
}

type SpecialForm interface {
	Value
	Execute(Env, []Value) (Value, error)
	IsPure() bool
}

type Macro interface {
	Value
	Expand([]Value) (Value, error)
	IsPure() bool
}

type Callable interface {
	Value
	Call([]Value) (Value, error)
	IsPure() bool
}

type Env interface {
	Bind(k string, v Value)
	BindRoot(k string, v Value)
	Lookup(k string) (Value, bool)
	MarkPure()
	IsInPureContext() bool
}

type Numeric interface {
	Value
	AsInt64() (int64, bool)
	AsDouble() (float64, bool)
}
