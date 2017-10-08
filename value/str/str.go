package str

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinarvk/heisenlisp/hashcode"
	"github.com/steinarvk/heisenlisp/types"
)

const TypeName = "string"

var (
	metricNewString = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "new_string",
			Help:      "New string values created",
		},
	)

	metricStringEqualityComparisons = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "string_equality_comparisons",
			Help:      "String equality comparisons made",
		},
	)

	metricStringConversionsToNative = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "string_conversions_to_native",
			Help:      "Conversions of string values to native Go strings",
		},
	)
)

func init() {
	prometheus.MustRegister(metricNewString)
	prometheus.MustRegister(metricStringEqualityComparisons)
	prometheus.MustRegister(metricStringConversionsToNative)
}

type stringValue string

func (s stringValue) AtomEquals(other types.Atom) bool {
	o, ok := other.(stringValue)
	if !ok {
		return false
	}
	metricStringEqualityComparisons.Inc()
	return o == s
}

func (s stringValue) String() string {
	return fmt.Sprintf("%q", string(s))
}

func (s stringValue) Eval(_ types.Env) (types.Value, error) { return s, nil }

func (s stringValue) Falsey() bool     { return s == "" }
func (_ stringValue) TypeName() string { return TypeName }

func ToString(v types.Value) (string, error) {
	rv, ok := v.(stringValue)
	if !ok {
		return "", errors.New("not a string")
	}
	metricStringConversionsToNative.Inc()
	return string(rv), nil
}

func New(s string) types.Value {
	metricNewString.Inc()
	return stringValue(s)
}

func (s stringValue) Hashcode() uint32 {
	return hashcode.Hash("str:", []byte(s))
}
