package cons

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/null"
)

const TypeName = "cons"

var (
	metricNewCons = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "new_cons",
			Help:      "New cons values created",
		},
	)
)

func init() {
	prometheus.MustRegister(metricNewCons)
}

var (
	notACons       = errors.New("not a cons")
	notAProperList = errors.New("not a proper list")
)

type consValue struct {
	car types.Value
	cdr types.Value

	h uint32

	conversionsToList int
	cachedListForm    []types.Value
}

func (_ *consValue) TypeName() string { return TypeName }
func (c *consValue) Falsey() bool     { return false }

func (c *consValue) String() string {
	xs := []string{}

	node := c
	for {
		xs = append(xs, node.car.String())
		next, ok := node.cdr.(*consValue)
		if !ok {
			if !null.IsNil(node.cdr) {
				xs = append(xs, ".")
				xs = append(xs, node.cdr.String())
			}
			break
		}

		node = next
	}

	return fmt.Sprintf("(%s)", strings.Join(xs, " "))
}

func NewChain(carItems []types.Value, cdr types.Value) types.Value {
	if len(carItems) == 0 {
		return cdr
	}

	if len(carItems) == 1 {
		return New(carItems[0], cdr)
	}

	return New(carItems[0], NewChain(carItems[1:], cdr))
}

func New(car, cdr types.Value) types.Value {
	if car == nil {
		car = null.Nil
	}
	if cdr == nil {
		cdr = null.Nil
	}

	metricNewCons.Inc()
	return &consValue{
		car: car,
		cdr: cdr,
		h:   hashcode.Hash("cons:", []byte(string(car.Hashcode())), []byte(string(cdr.Hashcode()))),
	}
}

func Decompose(v types.Value) (types.Value, types.Value, bool) {
	rv, ok := v.(*consValue)
	if !ok {
		return nil, nil, false
	}

	return rv.car, rv.cdr, true
}

func Car(v types.Value) (types.Value, error) {
	rv, ok := v.(*consValue)
	if !ok {
		return nil, notACons
	}
	return rv.car, nil
}

func Cdr(v types.Value) (types.Value, error) {
	rv, ok := v.(*consValue)
	if !ok {
		return nil, notACons
	}
	return rv.cdr, nil
}

func (c *consValue) Hashcode() uint32 {
	return c.h
}

func (c *consValue) Eval(e types.Env) (types.Value, error) {
	l, ok := c.asProperList()
	if !ok {
		return nil, fmt.Errorf("not a proper list")
	}

	if len(l) < 1 {
		return nil, errors.New("cannot evaluate empty list")
	}
	funcVal, err := l[0].Eval(e)
	if err != nil {
		return nil, err
	}

	unevaluatedParams := l[1:]

	specialForm, ok := funcVal.(types.SpecialForm)
	if ok {
		if !specialForm.IsPure() && e.IsInPureContext() {
			return nil, errors.New("impure call in pure context")
		}
		return specialForm.Execute(e, unevaluatedParams)
	}

	macro, ok := funcVal.(types.Macro)
	if ok {
		if !macro.IsPure() && e.IsInPureContext() {
			return nil, errors.New("impure call in pure context")
		}

		newForm, err := macro.Expand(unevaluatedParams)
		if err != nil {
			return nil, err
		}
		return newForm.Eval(e)
	}

	callable, ok := funcVal.(types.Callable)
	if !ok {
		return nil, fmt.Errorf("%q (%v) is not callable", l[0], funcVal)
	}
	if !callable.IsPure() && e.IsInPureContext() {
		return nil, errors.New("impure call in pure context")
	}

	n := len(unevaluatedParams)
	params := make([]types.Value, n, n)
	for i, unevaled := range unevaluatedParams {
		evaled, err := unevaled.Eval(e)
		if err != nil {
			return nil, err
		}
		params[i] = evaled
	}

	return callable.Call(params)
}

func (c *consValue) asProperList() ([]types.Value, bool) {
	if c.cachedListForm != nil {
		return c.cachedListForm, true
	}
	cache := false
	c.conversionsToList++
	if c.conversionsToList > 2 {
		cache = true
	}

	var rv []types.Value

	node := c
	for {
		rv = append(rv, node.car)
		if null.IsNil(node.cdr) {
			if cache {
				c.cachedListForm = rv
			}
			return rv, true
		}

		next, ok := node.cdr.(*consValue)
		if !ok {
			return nil, false
		}
		node = next
	}
}

func FromProperList(vs []types.Value) types.Value {
	return NewChain(vs, null.Nil)
}

func IsCons(v types.Value) bool {
	_, ok := v.(*consValue)
	return ok
}

func ToProperList(v types.Value) ([]types.Value, error) {
	if null.IsNil(v) {
		return nil, nil
	}

	rv, ok := v.(*consValue)
	if !ok {
		return nil, notACons
	}

	rrv, ok := rv.asProperList()
	if !ok {
		return nil, notAProperList
	}

	return rrv, nil
}
