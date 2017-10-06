package valuemap

import (
	"github.com/steinarvk/heisenlisp/cyclebreaker"
	"github.com/steinarvk/heisenlisp/types"
)

// A map from types.Value to interface{}.
// This is an intentionally bad implementation to establish the interface.

// To be replaced with a proper hash-based one later.

// Note that this cannot be a types.Value as it is not immutable.
// (A "frozen" version will later become a types.Value.)

// For setting, only must-equality is respected, anything else is considered separate.
// However, for getting, a separate kind of get is

type entry struct {
	key   types.Value
	value interface{}
}

type Map struct {
	entries []entry
}

func New() *Map {
	return &Map{}
}

func mustEqual(a, b types.Value) bool {
	tv, err := cyclebreaker.Equals(a, b)
	if err != nil {
		panic(err)
	}
	return tv == types.True
}

func mayEqual(a, b types.Value) bool {
	tv, err := cyclebreaker.Equals(a, b)
	if err != nil {
		panic(err)
	}
	return tv == types.True || tv == types.Maybe
}

func (m *Map) getPreviousAndMaybeSet(k types.Value, v interface{}) (interface{}, bool) {
	// linear search
	for _, entry := range m.entries {
		if mustEqual(entry.key, k) {
			oldVal := entry.value
			if v != nil {
				entry.value = v
			}
			return oldVal, true
		}
	}
	if v != nil {
		m.entries = append(m.entries, entry{k, v})
	}
	return nil, false
}

// SetAndGetPrevious sets the value of k to v, and returns the _previous_ value.
func (m *Map) SetAndGetPrevious(k types.Value, v interface{}) (interface{}, bool) {
	return m.getPreviousAndMaybeSet(k, v)
}

func (m *Map) GetMust(k types.Value, v interface{}) (interface{}, bool) {
	return m.getPreviousAndMaybeSet(k, nil)
}

func (m *Map) LookupMaybe(k types.Value, f func(types.Value, interface{}) bool) {
	for _, entry := range m.entries {
		if mayEqual(entry.key, k) {
			shortCircuit := f(entry.key, entry.value)
			if shortCircuit {
				return
			}
		}
	}
}
