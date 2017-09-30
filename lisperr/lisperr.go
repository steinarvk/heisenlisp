package lisperr

import (
	"errors"
	"fmt"
)

var DivisionByZero = errors.New("division by zero")

type UnboundVariable string

func (u UnboundVariable) Error() string {
	return fmt.Sprintf("unbound variable %q", string(u))
}
