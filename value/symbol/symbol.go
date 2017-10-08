package symbol

import (
	"errors"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/interntable"
	"github.com/steinarvk/heisenlisp/lisperr"
	"github.com/steinarvk/heisenlisp/types"
)

const (
	TypeName = "symbol"
)

var (
	metricNewSymbol = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "new_symbol",
			Help:      "New symbol values created",
		},
	)

	metricNewSymbolInterned = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "new_symbol_interned",
			Help:      "New unique symbol value interned",
		},
	)

	metricSymbolEqualityComparisons = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "symbol_equality_comparisons",
			Help:      "String equality comparisons made",
		},
	)

	metricSymbolConversionsToNative = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "symbol_conversions_to_native",
			Help:      "Conversions of symbol values to native Go strings",
		},
	)
)

func init() {
	prometheus.MustRegister(metricNewSymbol)
	prometheus.MustRegister(metricSymbolEqualityComparisons)
	prometheus.MustRegister(metricSymbolConversionsToNative)
}

var (
	symboltable = interntable.New()
)

type symbolValue uint32

func (i symbolValue) String() string {
	name, ok := symboltable.ToString(uint32(i))
	if !ok {
		return "#<invalid symbol>"
	}
	return name
}

func (i symbolValue) Eval(e types.Env) (types.Value, error) {
	val, ok := e.Lookup(uint32(i))
	if !ok {
		return nil, lisperr.UnboundVariable(i.String())
	}
	return val, nil
}

func (_ symbolValue) Falsey() bool { return false }

func (i symbolValue) AtomEquals(other types.Atom) bool {
	o, ok := other.(symbolValue)
	if !ok {
		return false
	}
	metricSymbolEqualityComparisons.Inc()
	return o == i
}

func (_ symbolValue) TypeName() string { return TypeName }

func Name(v types.Value) (string, error) {
	rv, ok := v.(symbolValue)
	if !ok {
		return "", errors.New("not a symbol")
	}
	metricSymbolConversionsToNative.Inc()
	return rv.String(), nil
}

func Id(v types.Value) (uint32, error) {
	rv, ok := v.(symbolValue)
	if !ok {
		return 0, errors.New("not a symbol")
	}
	return uint32(rv), nil
}

func IdOrPanic(v types.Value) uint32 {
	rv, err := Id(v)
	if err != nil {
		panic(err)
	}
	return rv
}

func StringToIdOrPanic(s string) uint32 {
	return IdOrPanic(New(s))
}

func New(s string) types.Value {
	s = strings.ToLower(s)
	metricNewSymbol.Inc()
	n, isNew := symboltable.ToInt(s)
	if isNew {
		metricNewSymbolInterned.Inc()
	}
	return symbolValue(n)
}

func Is(v types.Value) bool {
	_, ok := v.(symbolValue)
	return ok
}

func (s symbolValue) Hashcode() uint32 {
	return hashcode.Hash("sym:", []byte(string(s)))
}
