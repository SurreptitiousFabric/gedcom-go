package parser

import (
	"bytes"
	"testing"
)

func FuzzParseDoesNotPanic(f *testing.F) {
	seeds := [][]byte{
		[]byte("0 HEAD\n1 GEDC\n2 VERS 5.5\n0 TRLR"),
		[]byte("0 @I1@ INDI\n1 NAME John /Smith/\n0 TRLR"),
		[]byte("INVALID"),
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip("input too large")
		}
		p := NewParser()
		_, _ = p.Parse(bytes.NewReader(data))
	})
}
