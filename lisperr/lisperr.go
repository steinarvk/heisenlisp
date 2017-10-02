package lisperr

import (
	"errors"
	"fmt"

	"github.com/steinarvk/heisenlisp/types"
)

type UnexpectedValue struct {
	Expectation string
	Value       types.Value
}

func (u UnexpectedValue) Error() string {
	return fmt.Sprintf("unexpected value (%v): wanted %s", u.Value, u.Expectation)
}

var DivisionByZero = errors.New("division by zero")

type UnboundVariable string

func (u UnboundVariable) Error() string {
	return fmt.Sprintf("unbound variable %q", string(u))
}

type NotImplemented string

func (n NotImplemented) Error() string { return fmt.Sprintf("not implemented: %s", string(n)) }
