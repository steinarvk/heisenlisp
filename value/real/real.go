package real

import (
	"fmt"
	"math"

	"github.com/steinarvk/heisenlisp/types"
)

const TypeName = "floating-point"

type realValue float64

func (v realValue) AtomEquals(other types.Atom) bool {
	o, ok := other.(realValue)
	return ok && o == v
}

func (v realValue) String() string {
	return fmt.Sprintf("%f", float64(v))
}

func (v realValue) Eval(_ types.Env) (types.Value, error) { return v, nil }

func (v realValue) Falsey() bool { return v == 0 }

func (v realValue) TypeName() string { return TypeName }

func (v realValue) AsInt64() (int64, bool) {
	return 0, false
}

func (v realValue) AsDouble() (float64, bool) {
	return float64(v), true
}

func FromFloat64(x float64) types.Numeric {
	return realValue(x)
}

func IsNaN(v types.Value) bool {
	uv, ok := v.(realValue)
	if !ok {
		return false
	}
	return math.IsNaN(float64(uv))
}
