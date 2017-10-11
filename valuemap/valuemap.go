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
	m map[uint32][]entry
}

func New() *Map {
	return &Map{
		m: map[uint32][]entry{},
	}
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
	hc := k.Hashcode()
	if entries, present := m.m[hc]; present {
		// linear search
		for _, entry := range entries {
			if mustEqual(entry.key, k) {
				oldVal := entry.value
				if v != nil {
					entry.value = v
				}
				return oldVal, true
			}
		}
		if v != nil {
			m.m[hc] = append(entries, entry{k, v})
		}
	} else if v != nil {
		m.m[hc] = []entry{{k, v}}
	}
	return nil, false
}

// SetAndGetPrevious sets the value of k to v, and returns the _previous_ value.
func (m *Map) SetAndGetPrevious(k types.Value, v interface{}) (interface{}, bool) {
	return m.getPreviousAndMaybeSet(k, v)
}

func (m *Map) GetMust(k types.Value) (interface{}, bool) {
	return m.getPreviousAndMaybeSet(k, nil)
}

func (m *Map) LookupMaybe(k types.Value, f func(types.Value, interface{}) bool) {
	// todo: this could be optimised lots.
	// no actual application for this yet so it's not a huge priority.
	for _, entries := range m.m {
		for _, entry := range entries {
			if mayEqual(entry.key, k) {
				shortCircuit := f(entry.key, entry.value)
				if shortCircuit {
					return
				}
			}
		}
	}
}
