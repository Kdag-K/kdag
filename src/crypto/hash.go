package crypto

import (
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

// SHA256 returns the SHA256 hash of the data.
func SHA256(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	hash := hasher.Sum(nil)

	return hash
}

// Ripemd160 returns the Ripemd160 hash of the data.
func Ripemd160(bytes []byte) []byte {
	hasher := ripemd160.New()
	hasher.Write(bytes)

	return hasher.Sum(nil)
}

// SimpleHashFromTwoHashes returns the SHA256 hash of the concatenation of left
// and right data.
func SimpleHashFromTwoHashes(left []byte, right []byte) []byte {
	var hasher = sha256.New()
	hasher.Write(left)
	hasher.Write(right)

	return hasher.Sum(nil)
}
