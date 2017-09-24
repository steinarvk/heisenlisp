package numerics

import (
	"errors"
	"fmt"

	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/unknown"
	"github.com/steinarvk/heisenlisp/value/integer"
	"github.com/steinarvk/heisenlisp/value/real"
)

func wrapBinary(a, b types.Value, f func(a, b types.Value) (types.Value, error)) (types.Value, error) {
	av, ok1 := unknown.PossibleValues(a)
	bv, ok2 := unknown.PossibleValues(b)
	if !ok1 || !ok2 {
		// error?
		return unknown.FullyUnknown{}, nil
	}
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
		return unknown.NewMaybeAnyOf(rv)
	}

	return f(a, b)
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

func BinaryPlus(a, b types.Value) (types.Value, error) {
	if unknown.IsFullyUnknown(a) || unknown.IsFullyUnknown(b) {
		return unknown.FullyUnknown{}, nil
	}

	return wrapBinary(a, b, toBinaryTower(
		func(a, b int64) (types.Value, error) {
			return fromInt64(a + b), nil
		},
		func(a, b float64) (types.Value, error) {
			return real.FromFloat64(a + b), nil
		},
	))
}

func BinaryMinus(a, b types.Value) (types.Value, error) {
	if unknown.IsFullyUnknown(a) || unknown.IsFullyUnknown(b) {
		return unknown.FullyUnknown{}, nil
	}

	return wrapBinary(a, b, toBinaryTower(
		func(a, b int64) (types.Value, error) {
			return fromInt64(a - b), nil
		},
		func(a, b float64) (types.Value, error) {
			return real.FromFloat64(a - b), nil
		},
	))
}

func BinaryMultiply(a, b types.Value) (types.Value, error) {
	if unknown.IsFullyUnknown(a) || unknown.IsFullyUnknown(b) {
		return unknown.FullyUnknown{}, nil
	}

	return wrapBinary(a, b, toBinaryTower(
		func(a, b int64) (types.Value, error) {
			return fromInt64(a * b), nil
		},
		func(a, b float64) (types.Value, error) {
			return real.FromFloat64(a * b), nil
		},
	))
}

func BinaryDivision(a, b types.Value) (types.Value, error) {
	if unknown.IsFullyUnknown(a) || unknown.IsFullyUnknown(b) {
		return unknown.FullyUnknown{}, nil
	}

	return wrapBinary(a, b, toBinaryTower(
		func(a, b int64) (types.Value, error) {
			x := float64(a) / float64(b)
			return real.FromFloat64(x), nil
		},
		func(a, b float64) (types.Value, error) {
			return real.FromFloat64(a / b), nil
		},
	))
}

func Mod(a, b types.Value) (types.Value, error) {
	if unknown.IsFullyUnknown(a) || unknown.IsFullyUnknown(b) {
		return unknown.FullyUnknown{}, nil
	}

	return wrapBinary(a, b, toBinaryInt64(func(a, b int64) (types.Value, error) {
		if b == 0 {
			return nil, errors.New("division by zero")
		}
		return fromInt64(a % b), nil
	}))
}
