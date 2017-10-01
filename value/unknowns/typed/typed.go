// Package typed describes an unknown value of known or partially known type.
package typed

import (
	"fmt"
	"strings"

	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/typeset"
)

const TypeName = "unknown-of-type"

var _ types.Unknown = typedUnknown{}

type typedUnknown struct {
	ts *typeset.TypeSet
}

func (t typedUnknown) String() string {
	var xs []string
	for _, x := range t.ts.Slice() {
		xs = append(xs, x)
	}
	return fmt.Sprintf("#unknown-of-type(%s)", strings.Join(xs, " "))
}

func (t typedUnknown) Eval(_ types.Env) (types.Value, error) { return t, nil }
func (_ typedUnknown) Falsey() bool                          { return false }
func (_ typedUnknown) TypeName() string                      { return TypeName }

func (_ typedUnknown) HasNontypeInfo() bool { return false }

func (t typedUnknown) ActualTypeName() ([]string, bool) {
	return t.ts.Slice(), true
}

func (t typedUnknown) mayHaveType(name string) bool {
	return t.ts.Has(name)
}

func Is(v types.Value) bool {
	_, ok := v.(typedUnknown)
	return ok
}

func New(typenames ...string) types.Value {
	return typedUnknown{typeset.New(typenames...)}
}
