package intersection

import (
	"errors"
	"fmt"

	"github.com/steinarvk/heisenlisp/cyclebreaker"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/unknowns/anyof"
	"github.com/steinarvk/heisenlisp/value/unknowns/fullyunknown"
	"github.com/steinarvk/heisenlisp/value/unknowns/numinrange"
)

func linearStringSearch(xs []string, y string) bool {
	for _, x := range xs {
		if x == y {
			return true
		}
	}
	return false
}

func stringSlicesIntersect(xs, ys []string) bool {
	m := map[string]struct{}{}
	for _, x := range xs {
		m[x] = struct{}{}
	}
	for _, y := range ys {
		_, present := m[y]
		if present {
			return true
		}
	}
	return false
}

func unknownCanHaveValue(a types.Unknown, val types.Value) (bool, error) {
	// Can exclude on the right side: any form of unknown.

	valType := val.TypeName()

	if possibleTypename, typeKnown := a.ActualTypeName(); typeKnown {
		if !linearStringSearch(possibleTypename, valType) {
			return false, nil
		}
	}

	if !a.HasNontypeInfo() {
		// If there's no non-type info, then type intersection is sufficient.
		return true, nil
	}

	// Can exclude on left side: fullyunknown, typed.

	if vals, ok := anyof.PossibleValues(a); ok {
		for _, aVal := range vals {
			tv, err := cyclebreaker.Equals(aVal, val)
			if err != nil {
				return false, err
			}
			switch tv {
			case types.True:
				return true, nil
			case types.Maybe:
				return false, errors.New("impossible: unexpected uncertainty")
			}
		}

		return false, nil
	}

	// Can exclude on left side: fullyunknown, typed, anyof.
	if r, ok := numinrange.ToRange(a); ok {
		bNum, ok := val.(types.Numeric)
		if !ok {
			// Not a numeric, can't be contained in a range.
			return false, nil
		}

		return r.Contains(bNum), nil
	}

	return false, fmt.Errorf("unable to calculate intersection between unknown and value: %v and %v", a, val)
}

func unknownsIntersect(a, b types.Unknown) (bool, error) {
	if fullyunknown.Is(a) || fullyunknown.Is(b) {
		return true, nil
	}
	// Can now exclude on both sides: fullyunknown.

	if typesA, ok := a.ActualTypeName(); ok {
		if typesB, ok := b.ActualTypeName(); ok {
			if !stringSlicesIntersect(typesA, typesB) {
				return false, nil
			}
		}
	}
	if !a.HasNontypeInfo() || !b.HasNontypeInfo() {
		// intersection is sufficient since one side covers the entire type.
		return true, nil
	}
	// Can now exclude on both sides: fullyunknown, typed.

	if values, ok := anyof.PossibleValues(b); ok {
		for _, bVal := range values {
			rv, err := unknownCanHaveValue(a, bVal)
			if err != nil {
				return false, err
			}
			if rv {
				return true, nil
			}
		}
		return false, nil
	}
	if _, ok := anyof.PossibleValues(a); ok {
		return unknownsIntersect(b, a)
	}
	// Can now exclude on both sides: fullyunknown, typed, anyof.

	if rA, ok := numinrange.ToRange(a); ok {
		if rB, ok := numinrange.ToRange(b); ok {
			return rA.Intersection(rB) != nil, nil
		}

		// [indented]: can now exclude on both sides: fullyunknown, typed, anyof,
		//             as well as anything not an unknown, or not purely numeric.
		// There is nothing else such currently existing, so fail.
		return false, fmt.Errorf("unable to calculate intersection: %v and %v", a, b)
	}
	if _, ok := numinrange.ToRange(b); ok {
		return unknownsIntersect(b, a)
	}

	// Can now exclude on both sides: fullyunknown, typed, anyof, numinrange.

	return false, fmt.Errorf("unable to calculate intersection: %v and %v", a, b)
}

func intersectsUV(a types.Unknown, valB types.Value) (bool, error) {
	b, ok := valB.(types.Unknown)
	if !ok {
		return unknownCanHaveValue(a, valB)
	}
	return unknownsIntersect(a, b)
}

func Intersects(a, b types.Value) (bool, error) {
	unkA, ok := a.(types.Unknown)
	if ok {
		return intersectsUV(unkA, b)
	}
	unkB, ok := b.(types.Unknown)
	if ok {
		return intersectsUV(unkB, a)
	}

	tv, err := cyclebreaker.Equals(a, b)
	if err != nil {
		return false, err
	}
	switch tv {
	case types.True:
		return true, nil
	case types.False:
		return false, nil
	}
	panic("Equals() returned Maybe when comparing two non-unknown values")
}
