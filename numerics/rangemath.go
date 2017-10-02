package numerics

import (
	"github.com/steinarvk/heisenlisp/lisperr"
	"github.com/steinarvk/heisenlisp/numcmp"
	"github.com/steinarvk/heisenlisp/numrange"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/integer"
	"github.com/steinarvk/heisenlisp/value/unknowns/numinrange"
)

func rangeAdd(a, b *numrange.Range) (types.Value, error) {
	aHigh := a.UpperBound()
	bHigh := b.UpperBound()
	combinedHigh, err := BinaryPlus(aHigh, bHigh)
	if err != nil {
		return nil, err
	}

	combinedHighIncl := a.UpperBoundInclusive() && b.UpperBoundInclusive()

	aLow := a.LowerBound()
	bLow := b.LowerBound()
	combinedLow, err := BinaryPlus(aLow, bLow)
	if err != nil {
		return nil, err
	}

	combinedLowIncl := a.LowerBoundInclusive() && b.LowerBoundInclusive()

	// When we operate on two numerics, we should get a numeric back.
	combinedLowNum := combinedLow.(types.Numeric)
	combinedHighNum := combinedHigh.(types.Numeric)

	return numinrange.New(combinedLowNum, combinedHighNum, combinedLowIncl, combinedHighIncl, nil)
}

func rangeSub(a, b *numrange.Range) (types.Value, error) {
	aHigh := a.UpperBound()
	bHigh := b.UpperBound()

	aLow := a.LowerBound()
	bLow := b.LowerBound()

	// highest possibility: start with high, subtract low
	combinedHigh, err := BinaryMinus(aHigh, bLow)
	if err != nil {
		return nil, err
	}

	combinedHighIncl := a.UpperBoundInclusive() && b.LowerBoundInclusive()

	// lowest possibility: start with low, subtract high
	combinedLow, err := BinaryMinus(aLow, bHigh)
	if err != nil {
		return nil, err
	}

	combinedLowIncl := a.LowerBoundInclusive() && b.UpperBoundInclusive()

	// When we operate on two numerics, we should get a numeric back.
	combinedLowNum := combinedLow.(types.Numeric)
	combinedHighNum := combinedHigh.(types.Numeric)

	return numinrange.New(combinedLowNum, combinedHighNum, combinedLowIncl, combinedHighIncl, nil)
}

type valueWithInclusion struct {
	val       types.Numeric
	inclusive bool
}

func numericMulOrPanic(a, b types.Numeric) types.Numeric {
	result, err := BinaryMultiply(a, b)
	if err != nil {
		panic(err)
	}
	return result.(types.Numeric)
}

func numericDivOrPanic(a, b types.Numeric) types.Numeric {
	result, err := BinaryDivision(a, b)
	if err != nil {
		panic(err)
	}
	return result.(types.Numeric)
}

func rangeMul(a, b *numrange.Range) (types.Value, error) {
	vals := []valueWithInclusion{
		{
			numericMulOrPanic(a.LowerBound(), b.LowerBound()),
			a.LowerBoundInclusive() && b.LowerBoundInclusive(),
		},
		{
			numericMulOrPanic(a.LowerBound(), b.UpperBound()),
			a.LowerBoundInclusive() && b.UpperBoundInclusive(),
		},
		{
			numericMulOrPanic(a.UpperBound(), b.LowerBound()),
			a.UpperBoundInclusive() && b.LowerBoundInclusive(),
		},
		{
			numericMulOrPanic(a.UpperBound(), b.UpperBound()),
			a.UpperBoundInclusive() && b.UpperBoundInclusive(),
		},
	}

	// find the highest and the lowest of these numbers
	championLow := vals[0]
	championHigh := vals[0]

	for _, val := range vals[1:] {
		switch numcmp.CompareOrPanic(val.val, championLow.val) {
		case numcmp.Less:
			championLow = val
		case numcmp.Equal:
			if !championLow.inclusive && val.inclusive {
				championLow = val
			}
		}

		switch numcmp.CompareOrPanic(val.val, championHigh.val) {
		case numcmp.Greater:
			championHigh = val
		case numcmp.Equal:
			if !championHigh.inclusive && val.inclusive {
				championHigh = val
			}
		}
	}

	return numinrange.New(championLow.val, championHigh.val, championLow.inclusive, championHigh.inclusive, nil)
}

var zeroValueNumeric = integer.FromInt64(0).(types.Numeric)
var oneValueNumeric = integer.FromInt64(1).(types.Numeric)

func rangeDiv(a, b *numrange.Range) (types.Value, error) {
	if b.Contains(zeroValueNumeric) {
		return nil, lisperr.DivisionByZero
	}

	lowerBoundInverted := numericDivOrPanic(oneValueNumeric, b.LowerBound())
	upperBoundInverted := numericDivOrPanic(oneValueNumeric, b.UpperBound())

	bInv := numrange.New(upperBoundInverted, lowerBoundInverted, b.UpperBoundInclusive(), b.LowerBoundInclusive())

	return rangeMul(a, bInv)
}
