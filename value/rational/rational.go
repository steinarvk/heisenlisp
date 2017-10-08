package rational

import (
	"fmt"
	"math/big"

	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/types"
)

const TypeName = "rational"

var _ types.Numeric = rational{}

type rational struct {
	q *big.Rat
}

func (r rational) AtomEquals(other types.Atom) bool {
	o, ok := other.(rational)
	return ok && r.q.Cmp(o.q) == 0
}

func (r rational) String() string {
	return fmt.Sprintf("%v", r.q)
}

func (r rational) Eval(_ types.Env) (types.Value, error) { return r, nil }

func (r rational) Falsey() bool {
	return r.q.Cmp(new(big.Rat)) == 0
}

func (r rational) TypeName() string { return TypeName }

var maxInt63 = big.NewInt(9223372036854775807)

func (r rational) AsInt64() (int64, bool) {
	if !r.q.IsInt() {
		return 0, false
	}
	if new(big.Int).Abs(r.q.Num()).Cmp(maxInt63) > 0 {
		return 0, false
	}
	return r.q.Num().Int64(), true
}

func (r rational) AsDouble() (float64, bool) {
	rv, _ := r.q.Float64()
	return rv, true
}

func (r rational) AsBigint() (*big.Int, bool) {
	if !r.q.IsInt() {
		return nil, false
	}
	return r.q.Num(), true
}

func (r rational) AsBigrat() (*big.Rat, bool) {
	return r.q, true
}

func FromBig(q *big.Rat) types.Numeric {
	return rational{q}
}

func Parse(s string) (types.Numeric, error) {
	rv, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, fmt.Errorf("failed to parse %q as big rational", s)
	}
	return rational{rv}, nil
}

func (r rational) NumericRepresentationHashcode() uint32 {
	return hashcode.Hash("bigrat:", []byte(r.String()))
}

func (r rational) Hashcode() uint32 {
	// TODO: reduce to least number representation
	return r.NumericRepresentationHashcode()
}
