package optcons

import (
	"fmt"

	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/cons"
)

const TypeName = "optional-cons"

type optCons struct {
	car    types.Value
	cdr    types.Value
	mbCons types.Value
	h      uint32
}

func (o *optCons) Hashcode() uint32 {
	return o.h
}

var _ types.Unknown = &optCons{}

func New(car, cdr types.Value) types.Value {
	return &optCons{
		car:    car,
		cdr:    cdr,
		mbCons: cons.New(car, cdr),
		h:      hashcode.Hash("optcons:", []byte(string(car.Hashcode())), []byte(string(cdr.Hashcode()))),
	}
}

func (o *optCons) String() string {
	return fmt.Sprintf("#optional-cons(%v . %v)", o.car, o.cdr)
}

func (o *optCons) Eval(_ types.Env) (types.Value, error) { return o, nil }
func (o *optCons) TypeName() string                      { return TypeName }

func (o *optCons) HasNontypeInfo() bool { return true }

func (o *optCons) Falsey() bool {
	// todo? needs reworking, this would be maybe.
	return false
}

func (o *optCons) ActualTypeName() ([]string, bool) {
	unkCdr, ok := o.cdr.(types.Unknown)
	if !ok {
		tn := o.cdr.TypeName()
		if tn == cons.TypeName {
			return []string{cons.TypeName}, true
		}
		return []string{cons.TypeName, tn}, true
	}
	types, ok := unkCdr.ActualTypeName()
	if !ok {
		return nil, false
	}
	for _, x := range types {
		if x == cons.TypeName {
			return types, true
		}
	}
	return append(types, cons.TypeName), true
}

func Decompose(v types.Value) (types.Value, types.Value, bool) {
	rv, ok := v.(*optCons)
	if !ok {
		return nil, nil, false
	}

	return rv.car, rv.cdr, true
}

func Is(v types.Value) bool {
	_, ok := v.(*optCons)
	return ok
}

func OptConsValues(v types.Value) (types.Value, types.Value, bool) {
	oc, ok := v.(*optCons)
	if !ok {
		return nil, nil, false
	}
	return oc.mbCons, oc.cdr, true
}
