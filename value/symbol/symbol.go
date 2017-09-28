package symbol

import (
	"errors"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinarvk/heisenlisp/lisperr"
	"github.com/steinarvk/heisenlisp/types"
)

var (
	metricNewSymbol = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "new_symbol",
			Help:      "New symbol values created",
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

type symbolValue string

func (i symbolValue) String() string {
	return string(i)
}

func (i symbolValue) Eval(e types.Env) (types.Value, error) {
	n := string(i)
	val, ok := e.Lookup(n)
	if !ok {
		return nil, lisperr.UnboundVariable(n)
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

func (_ symbolValue) TypeName() string { return "symbol" }

func Name(v types.Value) (string, error) {
	rv, ok := v.(symbolValue)
	if !ok {
		return "", errors.New("not a symbol")
	}
	metricSymbolConversionsToNative.Inc()
	return string(rv), nil
}

func New(s string) types.Value {
	s = strings.ToLower(s)
	metricNewSymbol.Inc()
	return symbolValue(s)
}

func Is(v types.Value) bool {
	_, ok := v.(symbolValue)
	return ok
}
