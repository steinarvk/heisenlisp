package lisperr

import "fmt"

type UnboundVariable string

func (u UnboundVariable) Error() string {
	return fmt.Sprintf("unbound variable %q", string(u))
}
