package numtower

import (
	"errors"
	"math/big"

	"github.com/steinarvk/heisenlisp/types"
)

type BinaryTowerFunc struct {
	OnInt64s   func(int64, int64) (interface{}, error)
	OnBigints  func(*big.Int, *big.Int) (interface{}, error)
	OnBigrats  func(*big.Rat, *big.Rat) (interface{}, error)
	OnFloat64s func(float64, float64) (interface{}, error)
}

func (f BinaryTowerFunc) Call(a, b types.Numeric) (interface{}, error) {
	aInt, ok := a.AsInt64()
	bInt, ok2 := b.AsInt64()
	if ok && ok2 {
		return f.OnInt64s(aInt, bInt)
	}

	aBigint, ok := a.AsBigint()
	bBigint, ok2 := b.AsBigint()
	if ok && ok2 {
		return f.OnBigints(aBigint, bBigint)
	}

	aBigrat, ok := a.AsBigrat()
	bBigrat, ok2 := b.AsBigrat()
	if ok && ok2 {
		return f.OnBigrats(aBigrat, bBigrat)
	}

	aFloat, ok := a.AsDouble()
	bFloat, ok2 := b.AsDouble()
	if ok && ok2 {
		return f.OnFloat64s(aFloat, bFloat)
	}

	return nil, errors.New("reached end of numeric tower")
}
