package equality

import (
	"github.com/steinarvk/heisenlisp/numcmp"
	"github.com/steinarvk/heisenlisp/numerics"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/cons"
)

func AtomEquals(a, b types.Value) bool {
	av, aok := a.(types.Atom)
	bv, bok := b.(types.Atom)
	return aok && bok && av.AtomEquals(bv)
}

func ternaryAnd(a, b types.TernaryTruthValue) types.TernaryTruthValue {
	switch {
	case a == types.False || b == types.False:
		return types.False
	case a == types.True && b == types.True:
		return types.True
	default:
		return types.Maybe
	}
}

func Equals(a, b types.Value) (types.TernaryTruthValue, error) {
	if numerics.IsNumeric(a) && numerics.IsNumeric(b) {
		if numcmp.CompareOrPanic(a.(types.Numeric), b.(types.Numeric)) == numcmp.Equal {
			return types.True, nil
		}
		return types.False, nil
	}

	if AtomEquals(a, b) {
		return types.True, nil
	}

	if cons.IsCons(a) && cons.IsCons(b) {
		acar, acdr, _ := cons.Decompose(a)
		bcar, bcdr, _ := cons.Decompose(b)

		tv1, err := Equals(acar, bcar)
		if err != nil {
			return types.InvalidTernary, err
		}

		if tv1 == types.False {
			return types.False, nil
		}

		tv2, err := Equals(acdr, bcdr)
		if err != nil {
			return types.InvalidTernary, err
		}

		return ternaryAnd(tv1, tv2), nil
	}

	unkA, okA := a.(types.Unknown)
	if okA {
		ok, err := unkA.Intersects(b)
		if err != nil {
			return types.InvalidTernary, err
		}
		if ok {
			return types.Maybe, nil
		}
		return types.False, nil
	}

	unkB, okB := b.(types.Unknown)
	if okB {
		ok, err := unkB.Intersects(a)
		if err != nil {
			return types.InvalidTernary, err
		}
		if ok {
			return types.Maybe, nil
		}
		return types.False, nil
	}

	return types.False, nil
}
