package fullyunknown

import "github.com/steinarvk/heisenlisp/types"

type fullyUnknown struct{}

var Value fullyUnknown = fullyUnknown{}

func (_ fullyUnknown) String() string                        { return "#unknown" }
func (f fullyUnknown) Eval(_ types.Env) (types.Value, error) { return f, nil }
func (_ fullyUnknown) Falsey() bool                          { return false }
func (_ fullyUnknown) TypeName() string                      { return "unknown" }

func (_ fullyUnknown) Intersects(v types.Value) (bool, error) {
	// result is: v
	return true, nil
}

func Is(v types.Value) bool {
	_, ok := v.(fullyUnknown)
	return ok
}
