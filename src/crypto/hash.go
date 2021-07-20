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

// Rimp160 Returns hash: RIMP160( SHA256( data ) )
// Where possible, using RimpHash() should be a bit faster.
func Rimp160(b []byte) []byte {
	out := make([]byte, 20)
	rimpHash(b, out[:])

	return out[:]
}

func rimpHash(in []byte, out []byte) {
	sha := sha256.New()
	_, err := sha.Write(in)
	if err != nil {
		return
	}
	rim := ripemd160.New()
	_, err = rim.Write(sha.Sum(nil)[:])
	if err != nil {
		return
	}
	copy(out, rim.Sum(nil))
}
