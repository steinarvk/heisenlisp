package typeset

import (
	"sort"

	"github.com/steinarvk/heisenlisp/types"
)

type TypeSet struct {
	s []string
	m map[string]struct{}
}

func (s *TypeSet) Slice() []string {
	return s.s
}

func (s *TypeSet) Has(name string) bool {
	_, ok := s.m[name]
	return ok
}

func (s *TypeSet) HasAll(t *TypeSet) bool {
	for _, x := range t.Slice() {
		if !s.Has(x) {
			return false
		}
	}
	return true
}

func (s *TypeSet) IntersectsWith(v types.Value) bool {
	unk, ok := v.(types.Unknown)
	if !ok {
		return s.Has(v.TypeName())
	}

	theirTypes, ok := unk.ActualTypeName()
	if !ok {
		// They are fully unknown, so their intersection is equal to us.
		return true
	}

	for _, k := range theirTypes {
		if s.Has(k) {
			return true
		}
	}

	return false
}

func New(typenames ...string) *TypeSet {
	if len(typenames) == 0 {
		panic("new TypeSet without any possible types")
	}
	m := map[string]struct{}{}
	for _, t := range typenames {
		m[t] = struct{}{}
	}
	sort.Strings(typenames)
	return &TypeSet{
		s: typenames,
		m: m,
	}
}
