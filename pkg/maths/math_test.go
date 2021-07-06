package maths

import "testing"

func TestRoundUpPowerOfTwo(t *testing.T) {
	for i := 0; i < 100; i++ {
		n := RoundUpPowerOfTwo(uint64(i))
		t.Logf("%+v:%+v\n", i, n)
	}
}

func TestIsPowerOfTwo(t *testing.T) {
	for i := 0; i < 100; i++ {
		b := IsPowerOfTwo(uint64(i))
		t.Logf("%+v:%+v\n", i, b)
	}
}
