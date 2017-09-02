package expr

import (
	"fmt"
	"strings"
)

type Expr interface {
	String() string
}

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
