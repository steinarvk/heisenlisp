package expr

import (
	"fmt"
	"strings"
)

type Expr interface {
	String() string
}

type Bool bool

func (b Bool) String() string {
	if bool(b) {
		return "#true"
	}
	return "#false"
}

type AnyOf []Expr

func (a AnyOf) String() string {
	var xs []string
	for _, x := range a {
		xs = append(xs, x.String())
	}
	return fmt.Sprintf("#any-of(%s)", strings.Join(xs, " "))
}

type FullyUnknown struct{}

func (_ FullyUnknown) String() string { return "#unknown" }

type Identifier string

func (i Identifier) String() string {
	return string(i)
}

type Integer int64

func (i Integer) String() string {
	return fmt.Sprintf("%d", i)
}

type String string

func (s String) String() string {
	return fmt.Sprintf("%q", string(s))
}

type ListExpr []Expr

func (l ListExpr) String() string {
	var xs []string
	for _, x := range l {
		xs = append(xs, x.String())
	}
	return fmt.Sprintf("(%s)", strings.Join(xs, " "))
}
