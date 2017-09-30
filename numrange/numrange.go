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

// strictestUpperBound returns the strictest lower bound; which is
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
		return r.lowerBound, r.lowerBoundInclusive && o.lowerBoundInclusive
	}
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
