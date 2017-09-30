package logic

import (
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/unknown"
)

func TernaryAndValues(a, b types.Value) (types.TernaryTruthValue, error) {
	av, err := unknown.TruthValue(a)
	if err != nil {
		return types.InvalidTernary, nil
	}

	bv, err := unknown.TruthValue(b)
	if err != nil {
		return types.InvalidTernary, nil
	}

	return TernaryAnd(av, bv), nil
}

func TernaryOrValues(a, b types.Value) (types.TernaryTruthValue, error) {
	av, err := unknown.TruthValue(a)
	if err != nil {
		return types.InvalidTernary, nil
	}

	bv, err := unknown.TruthValue(b)
	if err != nil {
		return types.InvalidTernary, nil
	}

	return TernaryOr(av, bv), nil
}

func TernaryAnd(a, b types.TernaryTruthValue) types.TernaryTruthValue {
	switch {
	case a == types.False || b == types.False:
		return types.False
	case a == types.True && b == types.True:
		return types.True
	default:
		return types.Maybe
	}
}

func TernaryOr(a, b types.TernaryTruthValue) types.TernaryTruthValue {
	switch {
	case a == types.True || b == types.True:
		return types.True
	case a == types.False && b == types.False:
		return types.False
	default:
		return types.Maybe
	}
}
