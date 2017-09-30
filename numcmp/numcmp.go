// Package numcmp performs numeric comparisons not involving uncertainty.
package numcmp

import (
	"github.com/steinarvk/heisenlisp/numtower"
	"github.com/steinarvk/heisenlisp/types"
)

const (
	Less    int = -1
	Equal   int = 0
	Greater int = 1
)

var cmpNumerics = numtower.BinaryTowerFunc{
	OnInt64s: func(a, b int64) (interface{}, error) {
		diff := a - b
		switch {
		case diff < 0:
			return Less, nil
		case diff > 0:
			return Greater, nil
		default:
			return Equal, nil
		}
	},
	OnFloat64s: func(a, b float64) (interface{}, error) {
		diff := a - b
		switch {
		case diff < 0:
			return Less, nil
		case diff > 0:
			return Greater, nil
		default:
			return Equal, nil
		}
	},
}.Call

func Compare(a, b types.Numeric) (int, error) {
	valIf, err := cmpNumerics(a, b)
	if err != nil {
		return 0, err
	}
	return valIf.(int), nil
}

func CompareOrPanic(a, b types.Numeric) int {
	rv, err := Compare(a, b)
	if err != nil {
		panic(err)
	}
	return rv
}
