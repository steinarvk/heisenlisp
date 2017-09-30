// Package typed describes an unknown value of known or partially known type.
package typed

import (
	"fmt"
	"sort"
	"strings"

	"github.com/steinarvk/heisenlisp/types"
)

const TypeName = "unknown-of-type"

type typedUnknown struct {
	possibleTypeNames    []string
	possibleTypeNamesMap map[string]struct{}
}

func (t typedUnknown) String() string {
	var xs []string
	for _, x := range t.possibleTypeNames {
		xs = append(xs, x)
	}
	return fmt.Sprintf("#unknown-of-type(%s)", strings.Join(xs, " "))
}

func (t typedUnknown) Eval(_ types.Env) (types.Value, error) { return t, nil }
func (_ typedUnknown) Falsey() bool                          { return false }
func (_ typedUnknown) TypeName() string                      { return TypeName }

func (t typedUnknown) ActualTypeName() ([]string, bool) {
	return t.possibleTypeNames, true
}

func (t typedUnknown) mayHaveType(name string) bool {
	_, ok := t.possibleTypeNamesMap[name]
	return ok
}

func (t typedUnknown) Intersects(v types.Value) (bool, error) {
	unk, ok := v.(types.Unknown)
	if !ok {
		return t.mayHaveType(v.TypeName()), nil
	}

	theirTypes, ok := unk.ActualTypeName()
	if !ok {
		// They are fully unknown, so their intersection is equal to us.
		return true, nil
	}

	for _, k := range theirTypes {
		if t.mayHaveType(k) {
			return true, nil
		}
	}

	return false, nil
}

func Is(v types.Value) bool {
	_, ok := v.(typedUnknown)
	return ok
}

func New(typenames ...string) types.Value {
	if len(typenames) == 0 {
		panic("new unknown-of-type without any possible types")
	}
	m := map[string]struct{}{}
	for _, t := range typenames {
		m[t] = struct{}{}
	}
	sort.Strings(typenames)
	return typedUnknown{
		possibleTypeNames:    typenames,
		possibleTypeNamesMap: m,
	}
}
