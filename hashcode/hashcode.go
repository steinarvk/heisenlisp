package hashcode

import (
	"hash"
	"hash/fnv"
	"io"
)

func New() hash.Hash32 {
	return fnv.New32a()
}

func Hash(x string, data ...[]byte) uint32 {
	rv := New()
	io.WriteString(rv, x)
	for _, d := range data {
		rv.Write(d)
	}
	return rv.Sum32()
}
