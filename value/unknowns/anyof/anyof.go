package anyof

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/steinarvk/heisenlisp/cyclebreaker"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/boolean"
	"github.com/steinarvk/heisenlisp/value/unknowns/fullyunknown"
	"github.com/steinarvk/heisenlisp/value/unknowns/typed"
)

const TypeName = "any-of"

const (
	maxAnyOfElements = 100
)

type anyOf struct {
	vals  []types.Value
	types []string
}

var MaybeValue = anyOf{
	vals:  []types.Value{boolean.True, boolean.False},
	types: []string{boolean.TypeName},
}

func (a anyOf) isMaybe() bool {
	if len(a.vals) != 2 {
		return false
	}
	val1, err := boolean.ToBool(a.vals[0])
	if err != nil {
		return false
	}
	val2, err := boolean.ToBool(a.vals[1])
	if err != nil {
		return false
	}
	if val1 == true && val2 == false {
		return true
	}
	return val1 == false && val2 == true
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
func (_ anyOf) TypeName() string                      { return TypeName }

func (a anyOf) ActualTypeName() ([]string, bool) {
	return a.types, true
}

func (a anyOf) possibleValues() []types.Value {
	return a.vals
}

func (a anyOf) Intersects(v types.Value) (bool, error) {
	if fullyunknown.Is(v) {
		return true, nil
	}

	if typed.Is(v) {
		return v.(types.Unknown).Intersects(a)
	}

	xs, ok1 := PossibleValues(a)
	ys, ok2 := PossibleValues(v)
	if !ok1 || !ok2 {
		return false, fmt.Errorf("unable to calculate intersection of: %v and %v", a, v)
	}

	for _, x := range xs {
		for _, y := range ys {
			ternaryBool, err := cyclebreaker.Equals(x, y)
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

func newRaw(xs []types.Value) types.Value {
	rv := anyOf{}

	tps := map[string]struct{}{}

	addIfNew := func(singleValue types.Value) {
		// todo: do this more efficiently.
		for _, old := range rv.vals {
			if cyclebreaker.AtomEquals(old, singleValue) {
				return
			}
		}
		tps[singleValue.TypeName()] = struct{}{}
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

	var typesSlice []string
	for k := range tps {
		typesSlice = append(typesSlice, k)
	}
	sort.Strings(typesSlice)

	rv.types = typesSlice

	return rv
}

func NewOrPanic(xs []types.Value) types.Value {
	rv, err := New(xs)
	if err != nil {
		panic(err)
	}
	return rv
}

func New(xs []types.Value) (types.Value, error) {
	if len(xs) == 0 {
		return nil, errors.New("no options for any-of")
	}

	for _, x := range xs {
		if fullyunknown.Is(x) {
			return fullyunknown.Value, nil
		}
	}

	if len(xs) == 1 {
		return xs[0], nil
	}

	rv := newRaw(xs).(anyOf)

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
		return fullyunknown.Value, nil
	}

	return rv, nil
}

func PossibleValues(v types.Value) ([]types.Value, bool) {
	_, ok := v.(types.Unknown)
	if !ok {
		return []types.Value{v}, true
	}

	a, ok := v.(anyOf)
	if !ok {
		return nil, false
	}

	return a.possibleValues(), true
}

func IsMaybe(v types.Value) bool {
	rv, ok := v.(anyOf)
	if !ok {
		return false
	}
	return rv.isMaybe()
}
