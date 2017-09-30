package unknown

import (
	"fmt"

	"github.com/steinarvk/heisenlisp/types"

	"github.com/steinarvk/heisenlisp/value/unknowns/anyof"
)

var (
	MaybeValue = anyof.MaybeValue
)

func MayBeTruthy(v types.Value) (bool, error) {
	if !IsUncertain(v) {
		return !v.Falsey(), nil
	}

	if vs, ok := anyof.PossibleValues(v); ok {
		for _, sv := range vs {
			result, err := MayBeTruthy(sv)
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil
	}

	// unknown value not otherwise handled; assume yes.
	return true, nil
}

func MayBeFalsey(v types.Value) (bool, error) {
	if !IsUncertain(v) {
		return v.Falsey(), nil
	}

	if vs, ok := anyof.PossibleValues(v); ok {
		for _, sv := range vs {
			result, err := MayBeFalsey(sv)
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil
	}

	// unknown value not otherwise handled; assume yes.
	return true, nil
}

func TruthValue(v types.Value) (types.TernaryTruthValue, error) {
	mbTrue, err := MayBeTruthy(v)
	if err != nil {
		return types.InvalidTernary, err
	}
	mbFalse, err := MayBeFalsey(v)
	if err != nil {
		return types.InvalidTernary, err
	}
	switch {
	case !mbTrue && !mbFalse:
		return types.InvalidTernary, fmt.Errorf("value %v may neither be truthy nor falsey", v)
	case mbTrue && !mbFalse:
		return types.True, nil
	case !mbTrue && mbFalse:
		return types.False, nil
	default:
		return types.Maybe, nil
	}
}

func IsUncertain(v types.Value) bool {
	_, ok := v.(types.Unknown)
	return ok
}
