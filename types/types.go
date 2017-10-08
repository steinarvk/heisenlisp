package types

import "math/big"

type TernaryTruthValue string

const (
	False          = TernaryTruthValue("false")
	True           = TernaryTruthValue("true")
	Maybe          = TernaryTruthValue("maybe")
	InvalidTernary = TernaryTruthValue("invalid")
)

type Value interface {
	String() string
	Hashcode() uint32
	Falsey() bool
	TypeName() string
	Eval(e Env) (Value, error)
}

type Unknown interface {
	Value
	ActualTypeName() ([]string, bool)
	HasNontypeInfo() bool
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
	Bind(k uint32, v Value)
	BindRoot(k uint32, v Value)
	Lookup(k uint32) (Value, bool)
	MarkPure()
	IsInPureContext() bool
}

type Numeric interface {
	Value
	NumericRepresentationHashcode() uint32
	AsInt64() (int64, bool)
	AsBigint() (*big.Int, bool)
	AsBigrat() (*big.Rat, bool)
	AsDouble() (float64, bool)
}
