package database

import (
	"bytes"
	"fmt"
	"testing"
)

func TestUintBytesCoding(t *testing.T) {
	cases := map[uint64][]byte{
		0:                  []byte{0, 0, 0, 0, 0, 0, 0, 0},
		1:                  []byte{0, 0, 0, 0, 0, 0, 0, 1},
		2:                  []byte{0, 0, 0, 0, 0, 0, 0, 2},
		127:                []byte{0, 0, 0, 0, 0, 0, 0, 127},
		128:                []byte{0, 0, 0, 0, 0, 0, 0, 128},
		65:                 []byte{0, 0, 0, 0, 0, 0, 0, 65},
		0x6564636261605958: []byte{101, 100, 99, 98, 97, 96, 89, 88},
		0xFFFFFFFF:         []byte{0, 0, 0, 0, 0xff, 0xff, 0xff, 0xff},
		0xFFFFFFFFFFFFFFFF: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		0xFFFFFFFF00000000: []byte{0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0},
	}

	for in, expected := range cases {
		b := uint64ToBytes(in)
		if !bytes.Equal(b, expected) {
			fmt.Printf("uint64ToBytes() encoded %d wrong, got %v, expected %v\n", in, b, expected)
		}

		out := bytesToUint64(expected)
		if in != out {
			t.Fatalf("Failed to decode %v, got %d, expected %d\n", b, out, in)
		}
	}
}

// TestUintBytesCompare is a test to ansure that BoltDB will sort the keys
// correctly. We rely on the encoded result can be sorted by bytes.Compare().
func TestUintBytesCompare(t *testing.T) {
	cases := []struct {
		a        uint64
		b        uint64
		expected int
	}{
		{0, 0, 0},
		{0, 1, -1},
		{1, 1, 0},
		{10, 2, 1},
		{2, 10, -1},
		{22, 2, 1},
		{2, 22, -1},
		{65, 1, 1},
		{5111, 5110, 1},
		{5110, 5111, -1},
		{5110, 5110, 0},
		{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0},
		{0xFFFFFFFFFFFFFFFF, 0, 1},
		{0, 0xFFFFFFFFFFFFFFFF, -1},
		{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFE, 1},
		{0xFFFFFFFFFFFFFFFE, 0xFFFFFFFFFFFFFFFF, -1},
	}

	for _, c := range cases {
		ab := uint64ToBytes(c.a)
		bb := uint64ToBytes(c.b)
		compared := bytes.Compare(ab, bb)

		if compared != c.expected {
			t.Fatalf("Failed to compare %d and %d correctly when encoded\n", c.a, c.b)
		}
	}
}
