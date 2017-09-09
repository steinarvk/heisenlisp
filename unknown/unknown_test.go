package unknown

import (
	"testing"

	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/types"
)

type testcase struct {
	val  types.Value
	want TernaryTruthValue
}

func TestMaybeLogic(t *testing.T) {
	cases := []testcase{
		{expr.Bool(true), True},
		{expr.Bool(false), False},
		{NewMaybeAnyOf([]types.Value{
			expr.Bool(true),
			expr.Bool(false),
		}), Maybe},
	}
	for _, testcase := range cases {
		got, err := TruthValue(testcase.val)
		if err != nil {
			t.Errorf("TruthValue(%v) = err: %v", testcase.val, err)
			continue
		}
		if got != testcase.want {
			t.Errorf("TruthValue(%v) = %v want %v", testcase.val, got, testcase.want)
		}
	}
}
