package decoder

import (
	"bytes"
	"testing"
)

func FuzzDecodeDoesNotPanic(f *testing.F) {
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
		opts := DefaultOptions()
		opts.RecoverErrors = true
		_, _ = DecodeWithOptions(bytes.NewReader(data), opts)
	})
}
