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
func (_ FullyUnknown) TypeName() string                      { return "unknown" }

func (_ FullyUnknown) Intersects(v types.Value) (bool, error) {
	// result is: v
	return true, nil
}

func IsFullyUnknown(v types.Value) bool {
	_, ok := v.(FullyUnknown)
	return ok
}

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
func (a anyOf) Falsey() bool                          { return false }
func (_ anyOf) TypeName() string                      { return "any-of" }

func (a anyOf) possibleValues() []types.Value {
	return a.vals
}

func (a anyOf) Intersects(v types.Value) (bool, error) {
	if IsFullyUnknown(v) {
		return true, nil
	}

	xs, ok1 := PossibleValues(a)
	ys, ok2 := PossibleValues(v)
	if !ok1 || !ok2 {
		return false, fmt.Errorf("unable to calculate intersection of: %v and %v", a, v)
	}

	for _, x := range xs {
		for _, y := range ys {
			ternaryBool, err := expr.Equals(x, y)
			if err != nil {
				return false, err
			}
			switch ternaryBool {
			case types.False:
				break
			case types.Maybe:
				return true, nil
			case types.True:
				return true, nil
			}
		}
	}

	return false, nil
}

func PossibleValues(v types.Value) ([]types.Value, bool) {
	if !IsUncertain(v) {
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

	for _, x := range xs {
		if IsFullyUnknown(x) {
			return FullyUnknown{}, nil
		}
	}

	if len(xs) == 1 {
		return xs[0], nil
	}

	rv := NewAnyOf(xs).(anyOf)

	vals := rv.possibleValues()

	if len(vals) == 1 {
		return vals[0], nil
	}

	if len(vals) > maxAnyOfElements {
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
	if !IsUncertain(v) {
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
	if !IsUncertain(v) {
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
