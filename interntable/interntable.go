package interntable

import "sync"

type Table struct {
	mu          sync.Mutex
	stringToInt map[string]uint32
	intToString []string
}

func New() *Table {
	return &Table{
		mu:          sync.Mutex{},
		stringToInt: map[string]uint32{},
	}
}

func (t *Table) ToInt(s string) (uint32, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	rv, ok := t.stringToInt[s]
	if ok {
		return rv, false
	}

	index := len(t.intToString)
	t.intToString = append(t.intToString, s)
	result := uint32(index + 1)
	t.stringToInt[s] = result
	return result, true
}

func (t *Table) ToString(i uint32) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	index := int(i) - 1

	if index < len(t.intToString) {
		return t.intToString[index], true
	}
	return "", false
}
