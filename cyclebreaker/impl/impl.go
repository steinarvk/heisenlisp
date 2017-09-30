package impl

import (
	"github.com/steinarvk/heisenlisp/cyclebreaker"
	"github.com/steinarvk/heisenlisp/equality"
)

func init() {
	cyclebreaker.Equals = equality.Equals
	cyclebreaker.AtomEquals = equality.AtomEquals
}
