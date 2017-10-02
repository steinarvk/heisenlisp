package number

import (
	"fmt"
	"math"
	"strconv"

	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/integer"
	"github.com/steinarvk/heisenlisp/value/real"
)

func FromInt64(n int64) types.Numeric {
	return integer.FromInt64(n)
}

func FromFloat64(x float64) types.Numeric {
	// TODO? -- should we do this?
	// seems contrary to contagion.
	_, frac := math.Modf(x)
	if frac == 0 {
		return integer.FromInt64(int64(x))
	}

	return real.FromFloat64(x)
}

func FromString(s string) (types.Numeric, error) {
	if rv, err := integer.Parse(s); err == nil {
		return rv, nil
	}

	if x, err := strconv.ParseFloat(s, 64); err == nil {
		return FromFloat64(x), nil
	}

	return nil, fmt.Errorf("cannot parse %q as number", s)
}
