package numinrange

import (
	"fmt"

	"github.com/steinarvk/heisenlisp/numrange"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/typeset"
	"github.com/steinarvk/heisenlisp/value/integer"
	"github.com/steinarvk/heisenlisp/value/real"
	"github.com/steinarvk/heisenlisp/value/unknowns/anyof"
)

const TypeName = "number-in-range"

var _ types.Unknown = &numinrangeValue{}

var defaultTypeset = typeset.New(real.TypeName, integer.TypeName)

type numinrangeValue struct {
	ts *typeset.TypeSet
	r  *numrange.Range
}

func New(low, high types.Numeric, lowIncl, highIncl bool, typenames []string) (types.Value, error) {
	r := numrange.New(low, high, lowIncl, highIncl)
	ts := defaultTypeset
	if typenames != nil {
		ts = typeset.New(typenames...)
		if !defaultTypeset.HasAll(ts) {
			return nil, fmt.Errorf("invalid typeset %v for numeric range", ts)
		}
	}

	return &numinrangeValue{
		r:  r,
		ts: ts,
	}, nil
}

func (n *numinrangeValue) String() string {
	typeConstraint := ""
	if n.ts != defaultTypeset {
		typeConstraint = fmt.Sprintf(" %v", n.ts.Slice())
	}
	return fmt.Sprintf("#%s(%s%s)", TypeName, n.r.String(), typeConstraint)
}

func (n *numinrangeValue) Eval(_ types.Env) (types.Value, error) { return n, nil }
func (_ *numinrangeValue) Falsey() bool                          { return false }
func (_ *numinrangeValue) TypeName() string                      { return TypeName }

func (n *numinrangeValue) ActualTypeName() ([]string, bool) {
	return n.ts.Slice(), true
}

func (n *numinrangeValue) Intersects(v types.Value) (bool, error) {
	if !n.ts.IntersectsWith(v) {
		return false, nil
	}

	// We know there are intersections on a type level.
	// That's necessary but not sufficient.

	if values, ok := anyof.PossibleValues(v); ok {
		// This will cover any-of and single values.
		for _, val := range values {
			if !n.ts.IntersectsWith(val) {
				continue
			}
			// We assume that our own typeset will only contain numeric types.
			numericVal := val.(types.Numeric)
			if n.r.Contains(numericVal) {
				return true, nil
			}
		}

		return false, nil
	}

	if other, ok := v.(*numinrangeValue); ok {
		return n.r.Intersection(other.r) != nil, nil
	}

	panic(fmt.Sprintf("cannot check intersection between %v and %v", n, v))
}

func ToRange(v types.Value) (*numrange.Range, bool) {
	cast, ok := v.(*numinrangeValue)
	if !ok {
		return nil, false
	}
	return cast.r, true
}
