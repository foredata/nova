package bytex

import "testing"

func TestPool(t *testing.T) {
	maxSize := 1024 * 1024
	b1 := Alloc(10)
	b2 := Alloc(64)
	b3 := Alloc(65)
	b4 := Alloc(128)
	b5 := Alloc(maxSize - 1)
	b6 := Alloc(maxSize)
	b7 := Alloc(maxSize + 1)
	t.Log(len(b1), len(b2), len(b3), len(b4), len(b5), len(b6), len(b7))
	t.Log(Free(b1), Free(b2), Free(b3), Free(b4), Free(b5), Free(b6), Free(b7))
}
