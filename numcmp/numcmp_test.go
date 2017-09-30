package numcmp

import (
	"testing"

	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/integer"
)

func TestCompare(t *testing.T) {
	got := CompareOrPanic(integer.FromInt64(10).(types.Numeric), integer.FromInt64(20).(types.Numeric))
	if got != Less {
		t.Errorf("CompareOrPanic(10, 20) = %v want %v", got, Less)
	}
}
