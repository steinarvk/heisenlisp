package purity

import "strings"

func NameIsPure(s string) bool {
	return !strings.HasSuffix(s, "!")
}
