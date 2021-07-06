package metrics

// Inline and byte-free variant of hash/fnv(Fowler-Noll-Vo)'s fnv64a.

const (
	offset64 = 14695981039346656037
	prime64  = 1099511628211
)

// separatorByte is a byte that cannot occur in valid UTF-8 sequences and is
// used to separate label names, label values, and other strings from each other
// when calculating their combined hash value (aka signature aka fingerprint).
const separatorByte byte = 255

// hashNew initializies a new fnv64a hash value.
func hashNew() uint64 {
	return offset64
}

// hashAdd adds a string to a fnv64a hash value and add a seprator byte, returning the updated hash.
func hashAdd(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}

	h ^= uint64(separatorByte)
	h *= prime64

	return h
}
