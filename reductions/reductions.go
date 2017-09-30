package reductions

import (
	"github.com/steinarvk/heisenlisp/lisperr"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/cons"
	"github.com/steinarvk/heisenlisp/value/null"
	"github.com/steinarvk/heisenlisp/value/unknowns/anyof"
)

// how to stop exponential explosion here?
// since everything is immutable we could do some sort of memoisation
// which would likely often save us.

func FoldLeft(f func(a, b types.Value) (types.Value, error), initial types.Value, consish types.Value) (types.Value, error) {
	if null.IsNil(consish) {
		return initial, nil
	}

	car, cdr, ok := cons.Decompose(consish)
	if ok {
		val, err := f(initial, car)
		if err != nil {
			return nil, err
		}
		return FoldLeft(f, val, cdr)
	}

	conses, ok := anyof.PossibleValues(consish)
	if !ok {
		return nil, lisperr.UnexpectedValue{"cons or enumerable", consish}
	}

	var rv []types.Value

	for _, subconsish := range conses {
		// exploding! TODO
		val, err := FoldLeft(f, initial, subconsish)
		if err != nil {
			return nil, err
		}
		rv = append(rv, val)
	}

	return anyof.New(rv)
}

// Foldable returns whether something is a valid list or not. This is
// essentially the same as doing a fold with a trivial function that never
// returns an error, and throwing away the result.
func Foldable(consish types.Value) bool {
	// todo: handle things that _maybe_ a proper list

	f := func(_, _ types.Value) (types.Value, error) {
		return null.Nil, nil
	}
	_, err := FoldLeft(f, null.Nil, consish)
	// Only possible error should be UnexpectedValue{"cons or enumerable"}
	if err != nil {
		_, ok := err.(lisperr.UnexpectedValue)
		if ok {
			return false
		}
		panic(err)
	}
	return true
}
