package null

import "github.com/steinarvk/heisenlisp/types"

type nilValue struct{}

func IsNil(v types.Value) bool {
	_, ok := v.(nilValue)
	return ok
}

var (
	Nil = nilValue{}
)

func (_ nilValue) Falsey() bool                          { return true }
func (_ nilValue) String() string                        { return "nil" }
func (v nilValue) Eval(_ types.Env) (types.Value, error) { return v, nil }
func (v nilValue) AtomEquals(other types.Atom) bool {
	_, ok := other.(nilValue)
	return ok
}
func (_ nilValue) TypeName() string { return "nil" }
