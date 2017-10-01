package numrange

import (
	"strings"
	"testing"

	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/types"
)

func parseNumeric(e types.Env, s string) types.Numeric {
	rv, err := code.Run(e, "<testcase>", []byte(s))
	if err != nil {
		panic("invalid value")
	}
	return rv.(types.Numeric)
}

func parseSpec(e types.Env, s string) *Range {
	rv := &Range{}
	switch s[0] {
	case '[':
		rv.lowerBoundInclusive = true
	case '(':
		rv.lowerBoundInclusive = false
	default:
		panic("invalid first character")
	}
	switch s[len(s)-1] {
	case ']':
		rv.upperBoundInclusive = true
	case ')':
		rv.upperBoundInclusive = false
	default:
		panic("invalid last character")
	}

	s = s[1 : len(s)-1]
	comps := strings.Split(s, ",")
	if len(comps) != 2 {
		panic("expected two components")
	}

	if strings.TrimSpace(comps[0]) != "-inf" {
		rv.lowerBound = parseNumeric(e, comps[0])
	}

	if strings.TrimSpace(comps[1]) != "inf" {
		rv.upperBound = parseNumeric(e, comps[1])
	}
	return rv
}

type containsTestcase struct {
	spec  string
	value string
	want  bool
}

func TestRangeContains(t *testing.T) {
	e := builtin.NewRootEnv()

	testcases := []containsTestcase{
		{"[0, 10]", "5", true},
		{"[0, 10]", "0", true},
		{"[0, 10]", "10", true},
		{"[0, 10]", "-1", false},
		{"[0, 10]", "11", false},
		{"[0, 10]", "5.5", true},
		{"[0, 10]", "10.1", false},
		{"[0, 10]", "-0.1", false},
		{"(0, 10]", "0", false},
		{"(0, 10]", "10", true},
		{"[0, 10)", "0", true},
		{"[0, 10)", "10", false},
		{"[-inf, inf]", "5", true},
		{"[10,20]", "5", false},
		{"[10,20]", "15", true},
		{"[10,20]", "25", false},
	}

	for _, testcase := range testcases {
		r := parseSpec(e, testcase.spec)
		v := parseNumeric(e, testcase.value)
		got := r.Contains(v)
		if testcase.want != got {
			t.Errorf("%v.Contains(%v) = %v want %v (parsed from %q and %q)", r, v, got, testcase.want, testcase.spec, testcase.value)
		}
	}
}

type intersectionTestcase struct {
	spec1 string
	spec2 string
	want  string
}

func TestRangeIntersection(t *testing.T) {
	e := builtin.NewRootEnv()

	testcases := []intersectionTestcase{
		{"[0, 10]", "[-5,5]", "[0,5]"},
		{"[1, 5]", "[5,10]", "[5,5]"},
		{"[1, 5]", "[6,10]", "empty"},
		{"[1, 5]", "(5,10]", "empty"},
		{"[1, 5)", "[5,10]", "empty"},
	}

	for _, testcase := range testcases {
		a := parseSpec(e, testcase.spec1)
		b := parseSpec(e, testcase.spec2)
		got := a.Intersection(b).String()

		if testcase.want != got {
			t.Errorf("%v.Intersection(%v) = %v want %v (parsed from %q and %q)", a, b, got, testcase.want, testcase.spec1, testcase.spec2)
		}

		got = b.Intersection(a).String()

		if testcase.want != got {
			t.Errorf("%v.Intersection(%v) = %v want %v (parsed from %q and %q)", b, a, got, testcase.want, testcase.spec2, testcase.spec1)
		}
	}
}

func TestRangeRegressionSimple1(t *testing.T) {
	e := builtin.NewRootEnv()
	left := parseSpec(e, "[10,20]")
	right := parseSpec(e, "[0,0]")

	var got, want bool

	want = true
	got = left.otherRangeIsDisjointOnLowerSide(right)
	if got != want {
		t.Errorf("%v.otherRangeIsDisjointOnLowerSide(%v) = %v want %v", left, right, got, want)
	}

	want = false
	got = left.otherRangeIsDisjointOnUpperSide(right)
	if got != want {
		t.Errorf("%v.otherRangeIsDisjointOnLowerSide(%v) = %v want %v", left, right, got, want)
	}
}
