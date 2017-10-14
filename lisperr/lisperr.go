package lisperr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/steinarvk/heisenlisp/types"
)

type LispException struct {
	context []string
	value   types.Value
}

func NewException(v types.Value) LispException {
	return LispException{nil, v}
}

func (e LispException) Value() types.Value { return e.value }

func (e LispException) Error() string {
	rv := []string{"exception: "}
	for i := len(e.context) - 1; i >= 0; i-- {
		rv = append(rv, e.context[i], ": ")
	}
	return fmt.Sprintf("%s%v", strings.Join(rv, ""), e.value)
}

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

func Wrap(ctx string, err error) error {
	// if it's a LispException going in, it should be one going out as well
	if exc, ok := err.(LispException); ok {
		return LispException{
			context: append(exc.context, ctx),
			value:   exc.value,
		}
	}

	return fmt.Errorf("%s: %v", ctx, err)
}
