package null

import (
	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/types"
)

const TypeName = "nil"

type nilValue struct{}

var nilHash = hashcode.Hash("null:nil")

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
func (_ nilValue) TypeName() string { return TypeName }

func (_ nilValue) Hashcode() uint32 { return nilHash }
