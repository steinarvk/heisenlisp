package number

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/integer"
	"github.com/steinarvk/heisenlisp/value/rational"
	"github.com/steinarvk/heisenlisp/value/real"
)

func FromInt64(n int64) types.Numeric {
	return integer.FromInt64(n)
}

func FromFloat64(x float64) types.Numeric {
	return real.FromFloat64(x)
}

func FromBigInt(n *big.Int) types.Numeric {
	rv := integer.FromBig(n)

	if n, ok := rv.AsInt64(); ok {
		return integer.FromInt64(n)
	}

	return rv
}

func FromBigRat(n *big.Rat) types.Numeric {
	rv := rational.FromBig(n)

	if n, ok := rv.AsInt64(); ok {
		return integer.FromInt64(n)
	}

	if n, ok := rv.AsBigint(); ok {
		return integer.FromBig(n)
	}

	return rv
}

func FromString(s string) (types.Numeric, error) {
	if rv, err := integer.Parse(s); err == nil {
		return rv, nil
	}

	if x, err := strconv.ParseFloat(s, 64); err == nil {
		return FromFloat64(x), nil
	}

	if rv, err := rational.Parse(s); err == nil {
		return rv, nil
	}

	return nil, fmt.Errorf("cannot parse %q as number", s)
}
