package dedupe

import (
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/valuemap"
)

type Deduper struct {
	m  *valuemap.Map
	xs []types.Value
}

func New() *Deduper {
	return &Deduper{
		m: valuemap.New(),
	}
}

func (d *Deduper) Add(x types.Value) bool {
	_, seenBefore := d.m.SetAndGetPrevious(x, true)
	if !seenBefore {
		d.xs = append(d.xs, x)
	}
	return !seenBefore
}

func (d *Deduper) Slice() []types.Value {
	return d.xs
}

func Uniq(xs []types.Value) []types.Value {
	rv := New()
	for _, x := range xs {
		rv.Add(x)
	}
	return rv.Slice()
}
