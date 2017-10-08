package integer

import (
	"fmt"
	"math/big"

	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/types"
)

var _ types.Numeric = bigintValue{}

type bigintValue struct {
	n *big.Int
}

func (r bigintValue) AtomEquals(other types.Atom) bool {
	o, ok := other.(bigintValue)
	return ok && r.n.Cmp(o.n) == 0
}

func (r bigintValue) String() string {
	return fmt.Sprintf("%v", r.n)
}

func (r bigintValue) Eval(_ types.Env) (types.Value, error) { return r, nil }

func (r bigintValue) Falsey() bool {
	return r.n.Cmp(new(big.Int)) == 0
}

func (r bigintValue) TypeName() string { return TypeName }

var maxInt63 = big.NewInt(9223372036854775807)

func (r bigintValue) AsInt64() (int64, bool) {
	if new(big.Int).Abs(r.n).Cmp(maxInt63) > 0 {
		return 0, false
	}
	return r.n.Int64(), true
}

func (r bigintValue) AsDouble() (float64, bool) {
	val, _ := new(big.Rat).SetInt(r.n).Float64()
	return val, true
}

func (r bigintValue) AsBigint() (*big.Int, bool) {
	return r.n, true
}

func (r bigintValue) AsBigrat() (*big.Rat, bool) {
	return new(big.Rat).SetInt(r.n), true
}

func FromBig(n *big.Int) types.Numeric {
	return bigintValue{n}
}

func ParseBig(s string) (types.Numeric, error) {
	rv, ok := new(big.Int).SetString(s, 0)
	if !ok {
		return nil, fmt.Errorf("failed to parse %q as big integer", s)
	}
	return bigintValue{rv}, nil
}

func (r bigintValue) NumericRepresentationHashcode() uint32 {
	return hashcode.Hash("bigint:", []byte(r.String()))
}

func (r bigintValue) Hashcode() uint32 {
	// TODO: reduce to least number representation
	return r.NumericRepresentationHashcode()
}
