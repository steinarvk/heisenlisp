// Package numcmp performs numeric comparisons not involving uncertainty.
package numcmp

import (
	"math/big"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinarvk/heisenlisp/numtower"
	"github.com/steinarvk/heisenlisp/types"
)

var (
	metricNumericComparisons = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "numeric_comparisons",
			Help:      "Numeric comparisons made",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(metricNumericComparisons)
}

const (
	Less    int = -1
	Equal   int = 0
	Greater int = 1
)

var cmpNumerics = numtower.BinaryTowerFunc{
	OnInt64s: func(a, b int64) (interface{}, error) {
		metricNumericComparisons.WithLabelValues("int64").Inc()
		diff := a - b
		switch {
		case diff < 0:
			return Less, nil
		case diff > 0:
			return Greater, nil
		default:
			return Equal, nil
		}
	},
	OnBigints: func(a, b *big.Int) (interface{}, error) {
		metricNumericComparisons.WithLabelValues("bigint").Inc()
		return a.Cmp(b), nil
	},
	OnBigrats: func(a, b *big.Rat) (interface{}, error) {
		metricNumericComparisons.WithLabelValues("bigrat").Inc()
		return a.Cmp(b), nil
	},
	OnFloat64s: func(a, b float64) (interface{}, error) {
		metricNumericComparisons.WithLabelValues("float64").Inc()
		diff := a - b
		switch {
		case diff < 0:
			return Less, nil
		case diff > 0:
			return Greater, nil
		default:
			return Equal, nil
		}
	},
}.Call

func Compare(a, b types.Numeric) (int, error) {
	valIf, err := cmpNumerics(a, b)
	if err != nil {
		return 0, err
	}
	return valIf.(int), nil
}

func CompareOrPanic(a, b types.Numeric) int {
	rv, err := Compare(a, b)
	if err != nil {
		panic(err)
	}
	return rv
}
