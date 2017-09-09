package unknown

import (
	"fmt"
	"strings"

	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/types"
)

const (
	maxAnyOfElements = 20
)

type FullyUnknown struct{}

func (_ FullyUnknown) String() string                        { return "#unknown" }
func (f FullyUnknown) Eval(_ types.Env) (types.Value, error) { return f, nil }
func (_ FullyUnknown) Falsey() bool                          { return false }
func (_ FullyUnknown) Uncertain() bool                       { return true }
func (_ FullyUnknown) TypeName() string                      { return "unknown" }

type anyOf struct {
	vals []types.Value
}

func (a anyOf) String() string {
	var xs []string
	for _, x := range a.possibleValues() {
		xs = append(xs, x.String())
	}
	return fmt.Sprintf("#any-of(%s)", strings.Join(xs, " "))
}

func (a anyOf) Eval(_ types.Env) (types.Value, error) { return a, nil }
func (a anyOf) Uncertain() bool                       { return true }
func (a anyOf) Falsey() bool                          { return false }
func (_ anyOf) TypeName() string                      { return "any-of" }

func (a anyOf) possibleValues() []types.Value {
	return a.vals
}

func PossibleValues(v types.Value) ([]types.Value, bool) {
	if !v.Uncertain() {
		return []types.Value{v}, true
	}

	a, ok := v.(anyOf)
	if !ok {
		return nil, false
	}

	return a.possibleValues(), true
}

func NewAnyOf(xs []types.Value) types.Value {
	rv := anyOf{}

	// todo: do this more efficiently. n

	addIfNew := func(singleValue types.Value) {
		for _, old := range rv.vals {
			if expr.AtomEquals(old, singleValue) {
				return
			}
		}
		rv.vals = append(rv.vals, singleValue)
	}

	for _, x := range xs {
		vals, ok := PossibleValues(x)
		if ok {
			for _, val := range vals {
				addIfNew(val)
			}
			continue
		}
		addIfNew(x)
	}

	return rv
}

func NewMaybeAnyOf(xs []types.Value) types.Value {
	rv := NewAnyOf(xs).(anyOf)
	if len(rv.possibleValues()) > maxAnyOfElements {
		return FullyUnknown{}
	}
	return rv
}

func MayBeTruthy(v types.Value) (bool, error) {
	if !v.Uncertain() {
		return !v.Falsey(), nil
	}

	if vs, ok := PossibleValues(v); ok {
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
	if !v.Uncertain() {
		return v.Falsey(), nil
	}

	if vs, ok := PossibleValues(v); ok {
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

type TernaryTruthValue string

const (
	False   = TernaryTruthValue("false")
	True    = TernaryTruthValue("true")
	Maybe   = TernaryTruthValue("maybe")
	invalid = TernaryTruthValue("invalid")
)

func TruthValue(v types.Value) (TernaryTruthValue, error) {
	mbTrue, err := MayBeTruthy(v)
	if err != nil {
		return invalid, err
	}
	mbFalse, err := MayBeFalsey(v)
	if err != nil {
		return invalid, err
	}
	switch {
	case !mbTrue && !mbFalse:
		return invalid, fmt.Errorf("value %v may neither be truthy nor falsey", v)
	case mbTrue && !mbFalse:
		return True, nil
	case !mbTrue && mbFalse:
		return False, nil
	default:
		return Maybe, nil
	}
}
