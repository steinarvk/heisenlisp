package integer

// todo make type itself private

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/steinarvk/heisenlisp/types"
)

const TypeName = "integer"

type integer int64

func (i integer) AtomEquals(other types.Atom) bool {
	o, ok := other.(integer)
	return ok && o == i
}

func (i integer) String() string {
	return fmt.Sprintf("%d", i)
}

func (i integer) Eval(_ types.Env) (types.Value, error) { return i, nil }

func (i integer) Falsey() bool { return i == 0 }

func (_ integer) TypeName() string { return TypeName }

func (i integer) AsInt64() (int64, bool)    { return int64(i), true }
func (i integer) AsDouble() (float64, bool) { return float64(i), true }

func FromInt(v int) types.Value { return FromInt64(int64(v)) }

func FromInt64(v int64) types.Value {
	return integer(v)
}

func ToInt64(v types.Value) (int64, error) {
	rv, ok := v.(integer)
	if !ok {
		return 0, errors.New("not an integer")
	}
	return int64(rv), nil
}

func Parse(s string) (types.Value, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}
	return integer(n), nil
}
