package integer

import (
	"errors"
	"fmt"

	"github.com/steinarvk/heisenlisp/types"
)

type Integer int64

func (i Integer) AtomEquals(other types.Atom) bool {
	o, ok := other.(Integer)
	return ok && o == i
}

func (i Integer) String() string {
	return fmt.Sprintf("%d", i)
}

func (i Integer) Eval(_ types.Env) (types.Value, error) { return i, nil }

func (i Integer) Falsey() bool { return i == 0 }

func (_ Integer) TypeName() string { return "integer" }

func (i Integer) AsInt64() (int64, bool)    { return int64(i), true }
func (i Integer) AsDouble() (float64, bool) { return float64(i), true }

func FromInt(v int) types.Value { return FromInt64(int64(v)) }

func FromInt64(v int64) types.Value {
	return Integer(v)
}

func ToInt64(v types.Value) (int64, error) {
	rv, ok := v.(Integer)
	if !ok {
		return 0, errors.New("not an integer")
	}
	return int64(rv), nil
}
