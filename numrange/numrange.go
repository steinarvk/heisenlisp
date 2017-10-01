package numrange

import (
	"github.com/steinarvk/heisenlisp/numcmp"
	"github.com/steinarvk/heisenlisp/types"
)

type Range struct {
	lowerBound          types.Numeric
	upperBound          types.Numeric
	lowerBoundInclusive bool
	upperBoundInclusive bool
}

func (r *Range) LowerBound() types.Numeric { return r.lowerBound }
func (r *Range) UpperBound() types.Numeric { return r.upperBound }
func (r *Range) LowerBoundInclusive() bool { return r.lowerBoundInclusive }
func (r *Range) UpperBoundInclusive() bool { return r.upperBoundInclusive }

func NewSingleton(x types.Numeric) *Range {
	return &Range{
		upperBound:          x,
		lowerBound:          x,
		upperBoundInclusive: true,
		lowerBoundInclusive: true,
	}
}

func (r *Range) otherRangeMayBeGreater(other *Range) bool {
	whatIsMyLowerBound := numcmp.CompareOrPanic(r.LowerBound(), other.UpperBound())
	switch whatIsMyLowerBound {
	case numcmp.Less:
		// my lower bound is less than their upper bound, so they may be greater.
		return true
	case numcmp.Greater:
		// my lower bound is greater than their upper bound, so they may not be greater.
		return false
	case numcmp.Equal:
		// the bounds are equal. they may not be greater.
		return false
	default:
		panic("impossible")
	}
}

func (r *Range) otherRangeIsDisjointOnLowerSide(other *Range) bool {
	n := other.UpperBound()
	inclusive := r.lowerBoundInclusive && other.upperBoundInclusive

	if r.lowerBound == nil || n == nil {
		return false
	}
	result := numcmp.CompareOrPanic(n, r.lowerBound)
	if result == numcmp.Less {
		return true
	}
	if result == numcmp.Equal && !inclusive {
		return true
	}
	return false
}

func (r *Range) otherRangeIsDisjointOnUpperSide(other *Range) bool {
	n := other.LowerBound()
	inclusive := r.upperBoundInclusive && other.lowerBoundInclusive

	if r.upperBound == nil || n == nil {
		return false
	}
	result := numcmp.CompareOrPanic(n, r.upperBound)
	if result == numcmp.Greater {
		return true
	}
	if result == numcmp.Equal && !inclusive {
		return true
	}

	return false
}

func NewBelow(x types.Numeric, inclusive bool) *Range {
	return &Range{
		upperBound:          x,
		upperBoundInclusive: inclusive,
	}
}

func NewAbove(x types.Numeric, inclusive bool) *Range {
	return &Range{
		lowerBound:          x,
		lowerBoundInclusive: inclusive,
	}
}

func New(low, high types.Numeric, lowIncl, highIncl bool) *Range {
	// TODO? check valid / not empty
	return &Range{
		lowerBound:          low,
		upperBound:          high,
		lowerBoundInclusive: lowIncl,
		upperBoundInclusive: highIncl,
	}
}

func (r *Range) String() string {
	if r == nil {
		return "empty"
	}

	opener := "("
	closer := ")"
	if r.lowerBoundInclusive {
		opener = "["
	}
	if r.upperBoundInclusive {
		closer = "]"
	}

	s := opener
	if r.lowerBound != nil {
		s += r.lowerBound.String()
	} else {
		s += "-inf"
	}
	s += ","
	if r.upperBound != nil {
		s += r.upperBound.String()
	} else {
		s += "inf"
	}
	return s + closer
}

func (r *Range) Contains(n types.Numeric) bool {
	// say [10,100] testing 20
	if r.lowerBound != nil {
		// compare(10, 20) = Less [lower bound is less]
		switch numcmp.CompareOrPanic(r.lowerBound, n) {
		case numcmp.Greater:
			return false
		case numcmp.Equal:
			if !r.lowerBoundInclusive {
				return false
			}
		}
	}
	if r.upperBound != nil {
		switch numcmp.CompareOrPanic(n, r.upperBound) {
		case numcmp.Greater:
			return false
		case numcmp.Equal:
			if !r.upperBoundInclusive {
				return false
			}
		}
	}
	return true
}

// strictestLowerBound returns the strictest lower bound; which is
// a combination of a Numeric (nil means -inf) and a boolean
// meaning inclusive/exclusive.
func (r *Range) strictestLowerBound(o *Range) (types.Numeric, bool) {
	if r.lowerBound == nil {
		return o.lowerBound, o.lowerBoundInclusive
	}
	if o.lowerBound == nil {
		return r.lowerBound, r.lowerBoundInclusive
	}
	switch numcmp.CompareOrPanic(r.lowerBound, o.lowerBound) {
	case numcmp.Greater:
		return r.lowerBound, r.lowerBoundInclusive
	case numcmp.Less:
		return o.lowerBound, o.lowerBoundInclusive
	default:
		return r.lowerBound, r.lowerBoundInclusive && o.lowerBoundInclusive
	}
}

// strictestUpperBound returns the strictest upper bound; which is
// a combination of a Numeric (nil means -inf) and a boolean
// meaning inclusive/exclusive.
func (r *Range) strictestUpperBound(o *Range) (types.Numeric, bool) {
	if r.upperBound == nil {
		return o.upperBound, o.upperBoundInclusive
	}
	if o.upperBound == nil {
		return r.upperBound, r.upperBoundInclusive
	}
	switch numcmp.CompareOrPanic(r.upperBound, o.upperBound) {
	case numcmp.Greater:
		return o.upperBound, o.upperBoundInclusive
	case numcmp.Less:
		return r.upperBound, r.upperBoundInclusive
	default:
		return r.upperBound, r.upperBoundInclusive && o.upperBoundInclusive
	}
}

func (r *Range) IsSingleton() bool {
	if r.lowerBound == nil || r.upperBound == nil {
		return false
	}
	return numcmp.CompareOrPanic(r.lowerBound, r.upperBound) == numcmp.Equal
}

// Intersection computes the intersection of two ranges.
// Note that this can be nil!
func (r *Range) Intersection(o *Range) *Range {
	low, lowIncl := r.strictestLowerBound(o)
	high, highIncl := r.strictestUpperBound(o)

	switch numcmp.CompareOrPanic(low, high) {
	case numcmp.Equal:
		if !lowIncl || !highIncl {
			return nil
		}
		fallthrough
	case numcmp.Less:
		return &Range{
			lowerBound:          low,
			upperBound:          high,
			lowerBoundInclusive: lowIncl,
			upperBoundInclusive: highIncl,
		}
	case numcmp.Greater:
		return nil
	}
	panic("impossible")
}

type ComparisonResult struct {
	MayBeLeftLarger  bool
	MustBeLeftLarger bool

	MayBeEqual  bool
	MustBeEqual bool

	MayBeRightLarger  bool
	MustBeRightLarger bool
}

func Compare(left, right *Range) *ComparisonResult {
	if left.IsSingleton() && right.IsSingleton() {
		if numcmp.CompareOrPanic(left.lowerBound, right.lowerBound) == numcmp.Equal {
			// two singleton sets is the only way anything can be _must_ equal.
			return &ComparisonResult{
				MayBeEqual:  true,
				MustBeEqual: true,
			}
		}
	}

	// check for disjointness. two ways: left smaller, or left larger.
	if left.otherRangeIsDisjointOnLowerSide(right) {
		return &ComparisonResult{
			MayBeLeftLarger:  true,
			MustBeLeftLarger: true,
		}
	}
	if left.otherRangeIsDisjointOnUpperSide(right) {
		return &ComparisonResult{
			MayBeRightLarger:  true,
			MustBeRightLarger: true,
		}
	}

	// result will be uncertain.
	// we can exclude all the "must" options.
	// need to calculate what "may" options apply.

	mayBeEqual := left.Intersection(right) != nil
	mayBeRightGreater := left.otherRangeMayBeGreater(right)
	mayBeLeftGreater := right.otherRangeMayBeGreater(left)

	return &ComparisonResult{
		MayBeLeftLarger:  mayBeLeftGreater,
		MayBeRightLarger: mayBeRightGreater,
		MayBeEqual:       mayBeEqual,
	}
}
