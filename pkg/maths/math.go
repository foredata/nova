package maths

// http://graphics.stanford.edu/~seander/bithacks.html#DetermineIfPowerOf2
func RoundUpPowerOfTwo(v uint64) uint64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

func IsPowerOfTwo(v uint64) bool {
	return v > 0 && v&(v-1) == 0
}
