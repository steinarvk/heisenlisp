package fullyunknown

import (
	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/types"
)

const TypeName = "unknown"

type fullyUnknown struct{}

var _ types.Unknown = Value

var Value fullyUnknown = fullyUnknown{}

func (_ fullyUnknown) String() string                        { return "#unknown" }
func (f fullyUnknown) Eval(_ types.Env) (types.Value, error) { return f, nil }
func (_ fullyUnknown) Falsey() bool                          { return false }
func (_ fullyUnknown) TypeName() string                      { return TypeName }

func (_ fullyUnknown) HasNontypeInfo() bool { return false }

func (_ fullyUnknown) ActualTypeName() ([]string, bool) {
	return nil, false
}

func Is(v types.Value) bool {
	_, ok := v.(fullyUnknown)
	return ok
}

var unkHash = hashcode.Hash("fullyunknown:fullyunknown")

func (f fullyUnknown) Hashcode() uint32 { return unkHash }
