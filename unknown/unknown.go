package unknown

import (
	"errors"
	"fmt"
	"strings"

	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/types"
)

const (
	maxAnyOfElements = 100
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

func (a anyOf) isMaybe() bool {
	if len(a.vals) != 2 {
		return false
	}
	val1, ok1 := a.vals[0].(expr.Bool)
	val2, ok2 := a.vals[1].(expr.Bool)
	if !ok1 || !ok2 {
		return false
	}
	if val1 == expr.Bool(true) && val2 == expr.Bool(false) {
		return true
	}
	return val1 == expr.Bool(false) && val2 == expr.Bool(true)
}

func (a anyOf) String() string {
	if a.isMaybe() {
		return "maybe"
	}

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

var (
	MaybeValue = NewMaybeAnyOfOrPanic([]types.Value{
		expr.TrueValue, expr.FalseValue,
	})
)

func NewMaybeAnyOfOrPanic(xs []types.Value) types.Value {
	rv, err := NewMaybeAnyOf(xs)
	if err != nil {
		panic(err)
	}
	return rv
}

func NewMaybeAnyOf(xs []types.Value) (types.Value, error) {
	if len(xs) == 0 {
		return nil, errors.New("no options for any-of")
	}

	if len(xs) == 1 {
		return xs[0], nil
	}

	rv := NewAnyOf(xs).(anyOf)
	if len(rv.possibleValues()) > maxAnyOfElements {
		// past a certain limit we start discarding information to not allow the
		// work associated with keeping track of uncertainty to grow without bound.
		// note that returning a FullyUnknown is the last resort; other options
		// would be returning something with constrained type or value, e.g.
		// a numerically constrained value.
		return FullyUnknown{}, nil
	}

	return rv, nil
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
