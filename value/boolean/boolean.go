package boolean

import (
	"errors"

	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/types"
)

const (
	TypeName = "bool"
)

type boolValue bool

var (
	True  = boolValue(true)
	False = boolValue(false)
)

var (
	trueHash  = hashcode.Hash("bool:true")
	falseHash = hashcode.Hash("bool:false")
)

var (
	notABool = errors.New("not a bool")
)

func (b boolValue) AtomEquals(other types.Atom) bool {
	o, ok := other.(boolValue)
	return ok && o == b
}

func (b boolValue) Falsey() bool     { return !bool(b) }
func (_ boolValue) TypeName() string { return TypeName }

func (b boolValue) String() string {
	if bool(b) {
		return "true"
	}
	return "false"
}

func (b boolValue) Eval(_ types.Env) (types.Value, error) {
	return b, nil
}

func FromBool(b bool) types.Value {
	if b {
		return True
	}
	return False
}

func ToBool(v types.Value) (bool, error) {
	bv, ok := v.(boolValue)
	if !ok {
		return false, notABool
	}
	return bool(bv), nil
}

func (b boolValue) Hashcode() uint32 {
	if bool(b) {
		return trueHash
	}
	return falseHash
}
