package reductions

// Note: these aren't all reductions anymore, there's
// other things like map and filter.

import (
	"github.com/steinarvk/heisenlisp/lisperr"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/typeset"
	"github.com/steinarvk/heisenlisp/unknown"
	"github.com/steinarvk/heisenlisp/value/boolean"
	"github.com/steinarvk/heisenlisp/value/cons"
	"github.com/steinarvk/heisenlisp/value/null"
	"github.com/steinarvk/heisenlisp/value/unknowns/anyof"
	"github.com/steinarvk/heisenlisp/value/unknowns/optcons"
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

func FoldLeftWithOddTail(f func(a, b types.Value) (types.Value, error), handleOddTail func(a, b types.Value) (types.Value, error), initial types.Value, consish types.Value) (types.Value, error) {
	g := func(a, b types.Value) (types.Value, bool, error) {
		rv, err := f(a, b)
		return rv, false, err
	}
	return FoldLeftShortcircuitWithOddTail(g, handleOddTail, initial, consish)
}

func failOnOddTail(_, consish types.Value) (types.Value, error) {
	return nil, lisperr.UnexpectedValue{"cons or enumerable", consish}
}

func FoldLeftShortcircuit(f func(a, b types.Value) (types.Value, bool, error), initial types.Value, consish types.Value) (types.Value, error) {
	return FoldLeftShortcircuitWithOddTail(f, failOnOddTail, initial, consish)
}

func FoldLeftShortcircuitWithOddTail(f func(a, b types.Value) (types.Value, bool, error), handleOddTail func(a, b types.Value) (types.Value, error), initial types.Value, consish types.Value) (types.Value, error) {
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
		return FoldLeftShortcircuitWithOddTail(f, handleOddTail, val, cdr)
	}

	if !anyof.Is(consish) && !optcons.Is(consish) {
		return handleOddTail(initial, consish)
	}

	conses, _ := anyof.PossibleValues(consish)

	var rv []types.Value

	for _, subconsish := range conses {
		// exploding! TODO
		val, err := FoldLeftShortcircuitWithOddTail(f, handleOddTail, initial, subconsish)
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

func FilterReversed(f func(a types.Value) (types.Value, error), consish types.Value) (types.Value, error) {
	g := func(alreadyFolded, val types.Value) (types.Value, error) {
		result, err := f(val)
		if err != nil {
			return nil, err
		}
		tv, err := unknown.TruthValue(result)
		if err != nil {
			return nil, err
		}
		switch tv {
		case types.True:
			return cons.New(val, alreadyFolded), nil
		case types.False:
			return alreadyFolded, nil
		default:
			return optcons.New(val, alreadyFolded), nil
		}
	}
	return FoldLeft(g, null.Nil, consish)
}

func linearOnEach(f func(types.Value, bool), consish types.Value) (bool, error) {
	if null.IsNil(consish) {
		return true, nil
	}

	if car, cdr, ok := cons.Decompose(consish); ok {
		f(car, true)
		return linearOnEach(f, cdr)
	}

	if car, cdr, ok := optcons.Decompose(consish); ok {
		f(car, false)
		return linearOnEach(f, cdr)
	}

	_, ok := anyof.PossibleValues(consish)
	if !ok {
		return false, lisperr.UnexpectedValue{"cons or enumerable", consish}
	}

	// Linear iteration on a value with an any-of is impossible.
	return false, nil
}

type reversalEntry struct {
	value   types.Value
	creator func(a, b types.Value) types.Value
}

func resolveReversalEntries(entries []reversalEntry) types.Value {
	if len(entries) == 0 {
		return null.Nil
	}
	e := entries[len(entries)-1]
	rest := entries[:len(entries)-1]
	return e.creator(e.value, resolveReversalEntries(rest))
}

func Reversed(consish types.Value) (types.Value, error) {
	var inorder []reversalEntry

	f := func(val types.Value, alwaysPresent bool) {
		var entry reversalEntry
		if alwaysPresent {
			entry = reversalEntry{val, cons.New}
		} else {
			entry = reversalEntry{val, optcons.New}
		}
		inorder = append(inorder, entry)
	}

	ok, err := linearOnEach(f, consish)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, lisperr.NotImplemented("cannot reverse non-linearly-iterable list-like")
	}

	return resolveReversalEntries(inorder), nil
}

func Filter(f func(a types.Value) (types.Value, error), consish types.Value) (types.Value, error) {
	rv, err := FilterReversed(f, consish)
	if err != nil {
		return nil, err
	}
	return Reversed(rv)
}

func mapWithUncertainty(f func(types.Value) (types.Value, error), consish types.Value) (types.Value, error) {
	if null.IsNil(consish) {
		return null.Nil, nil
	}

	if car, cdr, ok := cons.Decompose(consish); ok {
		carMapped, err := f(car)
		if err != nil {
			return nil, err
		}
		cdrMapped, err := mapWithUncertainty(f, cdr)
		if err != nil {
			return nil, err
		}
		return cons.New(carMapped, cdrMapped), nil
	}

	if car, cdr, ok := optcons.Decompose(consish); ok {
		carMapped, err := f(car)
		if err != nil {
			return nil, err
		}
		cdrMapped, err := mapWithUncertainty(f, cdr)
		if err != nil {
			return nil, err
		}
		return optcons.New(carMapped, cdrMapped), nil
	}

	if !anyof.Is(consish) {
		return nil, lisperr.UnexpectedValue{"cons or enumerable", consish}
	}

	values, _ := anyof.PossibleValues(consish)

	var rv []types.Value

	for _, realConsish := range values {
		possible, err := mapWithUncertainty(f, realConsish)
		if err != nil {
			return nil, err
		}
		rv = append(rv, possible)
	}

	return anyof.New(rv)
}

func Map(f func(a types.Value) (types.Value, error), consish types.Value) (types.Value, error) {
	return mapWithUncertainty(f, consish)
}
