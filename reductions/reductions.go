package reductions

import (
	"github.com/steinarvk/heisenlisp/lisperr"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/typeset"
	"github.com/steinarvk/heisenlisp/value/boolean"
	"github.com/steinarvk/heisenlisp/value/cons"
	"github.com/steinarvk/heisenlisp/value/null"
	"github.com/steinarvk/heisenlisp/value/unknowns/anyof"
)

// how to stop exponential explosion here?
// since everything is immutable we could do some sort of memoisation
// which would likely often save us.

func FoldLeft(f func(a, b types.Value) (types.Value, error), initial types.Value, consish types.Value) (types.Value, error) {
	g := func(a, b types.Value) (types.Value, bool, error) {
		rv, err := f(a, b)
		return rv, false, err
	}
	return FoldLeftShortcircuit(g, initial, consish)
}

func FoldLeftShortcircuit(f func(a, b types.Value) (types.Value, bool, error), initial types.Value, consish types.Value) (types.Value, error) {
	if null.IsNil(consish) {
		return initial, nil
	}

	car, cdr, ok := cons.Decompose(consish)
	if ok {
		val, shortcircuit, err := f(initial, car)
		if err != nil {
			return nil, err
		}
		if shortcircuit {
			return val, nil
		}
		return FoldLeftShortcircuit(f, val, cdr)
	}

	conses, ok := anyof.PossibleValues(consish)
	if !ok {
		return nil, lisperr.UnexpectedValue{"cons or enumerable", consish}
	}

	var rv []types.Value

	for _, subconsish := range conses {
		// exploding! TODO
		val, err := FoldLeftShortcircuit(f, initial, subconsish)
		if err != nil {
			return nil, err
		}
		rv = append(rv, val)
	}

	return anyof.New(rv)
}

var (
	consTypeset = typeset.New(null.TypeName, cons.TypeName)
)

// Foldable returns whether something is a valid list or not. This is
// essentially the same as doing a fold with a trivial function that never
// returns an error, and throwing away the result.
func Foldable(consish types.Value) types.Value {
	if null.IsNil(consish) {
		return boolean.True
	}

	_, cdr, ok := cons.Decompose(consish)
	if ok {
		return Foldable(cdr)
	}

	conses, ok := anyof.PossibleValues(consish)
	if !ok {
		if consTypeset.IntersectsWith(consish) {
			return anyof.MaybeValue
		}
		return boolean.False
	}

	seenTrue := false
	seenFalse := false

	for _, subconsish := range conses {
		// exploding! TODO
		switch Foldable(subconsish) {
		case boolean.True:
			seenTrue = true
		case boolean.False:
			seenFalse = true
		default:
			return anyof.MaybeValue
		}

		if seenTrue && seenFalse {
			return anyof.MaybeValue
		}
	}

	switch {
	case seenTrue && !seenFalse:
		return boolean.True
	case seenFalse && !seenTrue:
		return boolean.False
	default:
		panic("impossible")
	}
}
