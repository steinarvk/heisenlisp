package numtower

import (
	"errors"

	"github.com/steinarvk/heisenlisp/types"
)

type BinaryTowerFunc struct {
	OnInt64s   func(int64, int64) (interface{}, error)
	OnFloat64s func(float64, float64) (interface{}, error)
}

func (f BinaryTowerFunc) Call(a, b types.Numeric) (interface{}, error) {
	aInt, ok := a.AsInt64()
	bInt, ok2 := b.AsInt64()
	if ok && ok2 {
		return f.OnInt64s(aInt, bInt)
	}

	aFloat, ok := a.AsDouble()
	bFloat, ok2 := b.AsDouble()
	if ok && ok2 {
		return f.OnFloat64s(aFloat, bFloat)
	}

	return nil, errors.New("reached end of numeric tower")
}
