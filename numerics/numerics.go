package numerics

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/steinarvk/heisenlisp/lisperr"
	"github.com/steinarvk/heisenlisp/number"
	"github.com/steinarvk/heisenlisp/numcmp"
	"github.com/steinarvk/heisenlisp/numrange"
	"github.com/steinarvk/heisenlisp/numtower"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/unknown"
	"github.com/steinarvk/heisenlisp/value/boolean"
	"github.com/steinarvk/heisenlisp/value/integer"

	"github.com/steinarvk/heisenlisp/value/unknowns/anyof"
	"github.com/steinarvk/heisenlisp/value/unknowns/fullyunknown"
	"github.com/steinarvk/heisenlisp/value/unknowns/numinrange"
)

func wrappedWithRanges(onNumerics func(a, b types.Numeric) (interface{}, error), onRanges func(a, b *numrange.Range) (interface{}, error)) func(a, b types.Value) (interface{}, error) {
	return func(a, b types.Value) (interface{}, error) {
		return wrapRangeOrNumeric(a, b, onNumerics, onRanges)
	}
}

func wrapRangeOrNumeric(a, b types.Value, onNumerics func(a, b types.Numeric) (interface{}, error), onRanges func(a, b *numrange.Range) (interface{}, error)) (interface{}, error) {
	rangeA, rA := numinrange.ToRange(a)
	rangeB, rB := numinrange.ToRange(b)
	numericA, nA := a.(types.Numeric)
	numericB, nB := b.(types.Numeric)

	switch {
	case rA && rB: // both ranges
		return onRanges(rangeA, rangeB)
	case rA && nB:
		return onRanges(rangeA, numrange.NewSingleton(numericB))
	case nA && rB:
		return onRanges(numrange.NewSingleton(numericA), rangeB)
	case nA && nB:
		return onNumerics(numericA, numericB)
	case !rA && !rB:
		return nil, fmt.Errorf("neigher range nor numeric: %v", a)
	default:
		return nil, fmt.Errorf("neigher range nor numeric: %v", b)
	}
}

func wrapBinary(a, b types.Value, f func(a, b types.Value) (types.Value, error)) (types.Value, error) {
	av, ok1 := anyof.PossibleValues(a)
	bv, ok2 := anyof.PossibleValues(b)
	if ok1 && ok2 {
		if len(av) > 1 || len(bv) > 1 {
			var rv []types.Value
			for _, ra := range av {
				for _, rb := range bv {
					res, err := f(ra, rb)
					if err != nil {
						return nil, err
					}
					rv = append(rv, res)
				}
			}
			return anyof.New(rv)
		}
	}

	return f(a, b)
}

func castingToValue(f func(a, b types.Value) (interface{}, error)) func(a, b types.Value) (types.Value, error) {
	return func(a, b types.Value) (types.Value, error) {
		rv, err := f(a, b)
		if err != nil {
			return nil, err
		}
		return rv.(types.Value), nil
	}
}

func toBinaryNumericValues(f func(types.Numeric, types.Numeric) (interface{}, error)) func(types.Value, types.Value) (types.Value, error) {
	return func(a, b types.Value) (types.Value, error) {
		an, ok := a.(types.Numeric)
		if !ok {
			return nil, fmt.Errorf("not a number: %v", a)
		}

		bn, ok := b.(types.Numeric)
		if !ok {
			return nil, fmt.Errorf("not a number: %v", b)
		}

		rvIf, err := f(an, bn)
		if err != nil {
			return nil, err
		}

		return rvIf.(types.Value), nil
	}
}

func toBinaryNumeric(f func(types.Numeric, types.Numeric) (types.Value, error)) func(types.Value, types.Value) (types.Value, error) {
	return func(a, b types.Value) (types.Value, error) {
		an, ok := a.(types.Numeric)
		if !ok {
			return nil, fmt.Errorf("not a number: %v", a)
		}

		bn, ok := b.(types.Numeric)
		if !ok {
			return nil, fmt.Errorf("not a number: %v", b)
		}

		return f(an, bn)
	}
}

func toBinaryInt64(f func(int64, int64) (types.Value, error)) func(types.Value, types.Value) (types.Value, error) {
	return toBinaryNumeric(func(a, b types.Numeric) (types.Value, error) {
		an, ok := a.AsInt64()
		if !ok {
			return nil, fmt.Errorf("not an int64: %v", a)
		}

		bn, ok := b.AsInt64()
		if !ok {
			return nil, fmt.Errorf("not an int64: %v", b)
		}

		return f(an, bn)
	})
}

func toBinaryTower(onInts func(int64, int64) (types.Value, error), onDoubles func(float64, float64) (types.Value, error)) func(types.Value, types.Value) (types.Value, error) {
	return toBinaryNumeric(func(a, b types.Numeric) (types.Value, error) {
		aInt, ok := a.AsInt64()
		bInt, ok2 := b.AsInt64()
		if ok && ok2 {
			return onInts(aInt, bInt)
		}

		aFloat, ok := a.AsDouble()
		bFloat, ok2 := b.AsDouble()
		if ok && ok2 {
			return onDoubles(aFloat, bFloat)
		}

		return nil, fmt.Errorf("not convertible to common numeric type: %v and %v", a, b)
	})
}

func fromInt64(n int64) types.Value {
	return integer.FromInt64(n)
}

func areSmall(a, b int64) bool {
	return int64(int32(a)) == a && int64(int32(b)) == b
}

var binaryPlus func(a, b types.Value) (types.Value, error)

func init() {
	binaryPlus = castingToValue(wrappedWithRanges(numtower.BinaryTowerFunc{
		OnInt64s: func(a, b int64) (interface{}, error) {
			if areSmall(a, b) {
				return number.FromInt64(a + b), nil
			}
			rv := big.NewInt(a)
			rv.Add(rv, big.NewInt(b))
			return number.FromBigInt(rv), nil
		},
		OnBigints: func(a, b *big.Int) (interface{}, error) {
			return number.FromBigInt(new(big.Int).Add(a, b)), nil
		},
		OnBigrats: func(a, b *big.Rat) (interface{}, error) {
			return number.FromBigRat(new(big.Rat).Add(a, b)), nil
		},
		OnFloat64s: func(a, b float64) (interface{}, error) {
			return number.FromFloat64(a + b), nil
		},
	}.Call, func(a, b *numrange.Range) (interface{}, error) {
		return rangeAdd(a, b)
	}))
}

func BinaryPlus(a, b types.Value) (types.Value, error) {
	if fullyunknown.Is(a) || fullyunknown.Is(b) {
		return fullyunknown.Value, nil
	}

	return wrapBinary(a, b, binaryPlus)
}

var binaryMinus func(a, b types.Value) (types.Value, error)

func init() {
	binaryMinus = castingToValue(wrappedWithRanges(numtower.BinaryTowerFunc{
		OnInt64s: func(a, b int64) (interface{}, error) {
			if areSmall(a, b) {
				return number.FromInt64(a - b), nil
			}
			rv := big.NewInt(a)
			rv.Sub(rv, big.NewInt(b))
			return number.FromBigInt(rv), nil
		},
		OnBigints: func(a, b *big.Int) (interface{}, error) {
			return number.FromBigInt(new(big.Int).Sub(a, b)), nil
		},
		OnBigrats: func(a, b *big.Rat) (interface{}, error) {
			return number.FromBigRat(new(big.Rat).Sub(a, b)), nil
		},
		OnFloat64s: func(a, b float64) (interface{}, error) {
			return number.FromFloat64(a - b), nil
		},
	}.Call, func(a, b *numrange.Range) (interface{}, error) {
		return rangeSub(a, b)
	}))
}

func BinaryMinus(a, b types.Value) (types.Value, error) {
	if fullyunknown.Is(a) || fullyunknown.Is(b) {
		return fullyunknown.Value, nil
	}

	return wrapBinary(a, b, binaryMinus)
}

var binaryMultiply func(a, b types.Value) (types.Value, error)

func init() {
	binaryMultiply = castingToValue(wrappedWithRanges(numtower.BinaryTowerFunc{
		OnInt64s: func(a, b int64) (interface{}, error) {
			if areSmall(a, b) {
				return number.FromInt64(a * b), nil
			}
			rv := big.NewInt(a)
			rv.Mul(rv, big.NewInt(b))
			return number.FromBigInt(rv), nil
		},
		OnBigints: func(a, b *big.Int) (interface{}, error) {
			return number.FromBigInt(new(big.Int).Mul(a, b)), nil
		},
		OnBigrats: func(a, b *big.Rat) (interface{}, error) {
			return number.FromBigRat(new(big.Rat).Mul(a, b)), nil
		},
		OnFloat64s: func(a, b float64) (interface{}, error) {
			return number.FromFloat64(a * b), nil
		},
	}.Call, func(a, b *numrange.Range) (interface{}, error) {
		return rangeMul(a, b)
	}))
}

func BinaryMultiply(a, b types.Value) (types.Value, error) {
	if fullyunknown.Is(a) || fullyunknown.Is(b) {
		return fullyunknown.Value, nil
	}

	return wrapBinary(a, b, binaryMultiply)
}

var binaryDivision func(a, b types.Value) (types.Value, error)

func init() {
	binaryDivision = castingToValue(wrappedWithRanges(numtower.BinaryTowerFunc{
		OnInt64s: func(a, b int64) (interface{}, error) {
			if b == 0 {
				return nil, lisperr.DivisionByZero
			}
			return number.FromBigRat(big.NewRat(a, b)), nil
		},
		OnBigints: func(a, b *big.Int) (interface{}, error) {
			if new(big.Int).Cmp(b) == 0 {
				return nil, lisperr.DivisionByZero
			}
			result := new(big.Rat)
			result.SetInt(a)
			result.Quo(result, new(big.Rat).SetInt(b))
			return number.FromBigRat(result), nil
		},
		OnBigrats: func(a, b *big.Rat) (interface{}, error) {
			if new(big.Rat).Cmp(b) == 0 {
				return nil, lisperr.DivisionByZero
			}
			return number.FromBigRat(new(big.Rat).Quo(a, b)), nil
		},
		OnFloat64s: func(a, b float64) (interface{}, error) {
			if b == 0 {
				return nil, lisperr.DivisionByZero
			}
			return number.FromFloat64(a / b), nil
		},
	}.Call, func(a, b *numrange.Range) (interface{}, error) {
		return rangeDiv(a, b)
	}))
}

func BinaryDivision(a, b types.Value) (types.Value, error) {
	if fullyunknown.Is(a) || fullyunknown.Is(b) {
		return fullyunknown.Value, nil
	}

	return wrapBinary(a, b, binaryDivision)
}

func Mod(a, b types.Value) (types.Value, error) {
	if fullyunknown.Is(a) || fullyunknown.Is(b) {
		return fullyunknown.Value, nil
	}

	return wrapBinary(a, b, toBinaryInt64(func(a, b int64) (types.Value, error) {
		if b == 0 {
			return nil, errors.New("division by zero")
		}
		return fromInt64(a % b), nil
	}))
}

var numericLeq = castingToValue(wrappedWithRanges(func(a, b types.Numeric) (interface{}, error) {
	// <=
	return boolean.FromBool(numrange.NewBelow(b, true).Contains(a)), nil
}, func(a, b *numrange.Range) (interface{}, error) {
	result := numrange.Compare(a, b)
	switch {
	case result.MustBeRightLarger || result.MustBeEqual:
		return boolean.True, nil
	case result.MayBeRightLarger || result.MayBeEqual:
		return anyof.MaybeValue, nil
	default:
		return boolean.False, nil
	}
}))

func BinaryLeq(a, b types.Value) (types.Value, error) {
	if fullyunknown.Is(a) || fullyunknown.Is(b) {
		return unknown.MaybeValue, nil
	}

	return wrapBinary(a, b, numericLeq)
}

var numericLess = castingToValue(wrappedWithRanges(func(a, b types.Numeric) (interface{}, error) {
	// <
	return boolean.FromBool(numrange.NewBelow(b, false).Contains(a)), nil
}, func(a, b *numrange.Range) (interface{}, error) {
	result := numrange.Compare(a, b)
	switch {
	case result.MustBeRightLarger:
		return boolean.True, nil
	case result.MayBeRightLarger:
		return anyof.MaybeValue, nil
	default:
		return boolean.False, nil
	}
}))

func BinaryLess(a, b types.Value) (types.Value, error) {
	if fullyunknown.Is(a) || fullyunknown.Is(b) {
		return unknown.MaybeValue, nil
	}

	return wrapBinary(a, b, numericLess)
}

var numericGeq = castingToValue(wrappedWithRanges(func(a, b types.Numeric) (interface{}, error) {
	// >=
	return boolean.FromBool(numrange.NewAbove(b, true).Contains(a)), nil
}, func(a, b *numrange.Range) (interface{}, error) {
	result := numrange.Compare(a, b)
	switch {
	case result.MustBeLeftLarger || result.MustBeEqual:
		return boolean.True, nil
	case result.MayBeLeftLarger || result.MayBeEqual:
		return anyof.MaybeValue, nil
	default:
		return boolean.False, nil
	}
}))

func BinaryGeq(a, b types.Value) (types.Value, error) {
	if fullyunknown.Is(a) || fullyunknown.Is(b) {
		return unknown.MaybeValue, nil
	}

	return wrapBinary(a, b, numericGeq)
}

var numericGreater = castingToValue(wrappedWithRanges(func(a, b types.Numeric) (interface{}, error) {
	// >
	return boolean.FromBool(numrange.NewAbove(b, false).Contains(a)), nil
}, func(a, b *numrange.Range) (interface{}, error) {
	result := numrange.Compare(a, b)
	switch {
	case result.MustBeLeftLarger:
		return boolean.True, nil
	case result.MayBeLeftLarger:
		return anyof.MaybeValue, nil
	default:
		return boolean.False, nil
	}
}))

func BinaryGreater(a, b types.Value) (types.Value, error) {
	if fullyunknown.Is(a) || fullyunknown.Is(b) {
		return unknown.MaybeValue, nil
	}

	return wrapBinary(a, b, numericGreater)
}

var numericEq = castingToValue(wrappedWithRanges(func(a, b types.Numeric) (interface{}, error) {
	// =
	eq := numcmp.CompareOrPanic(a, b) == numcmp.Equal
	return boolean.FromBool(eq), nil
}, func(a, b *numrange.Range) (interface{}, error) {
	result := numrange.Compare(a, b)
	switch {
	case result.MustBeEqual:
		return boolean.True, nil
	case result.MayBeEqual:
		return anyof.MaybeValue, nil
	default:
		return boolean.False, nil
	}
}))

func BinaryEq(a, b types.Value) (types.Value, error) {
	if fullyunknown.Is(a) || fullyunknown.Is(b) {
		return unknown.MaybeValue, nil
	}

	return wrapBinary(a, b, numericEq)
}

func IsNumeric(v types.Value) bool {
	_, ok := v.(types.Numeric)
	return ok
}
