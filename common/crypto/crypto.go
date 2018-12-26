package crypto

import (
	"github.com/lemoTestCoin/common"
	"github.com/lemoTestCoin/common/crypto/sha3"
	"github.com/lemoTestCoin/common/math"
	"math/big"
)

// lemochain 版本号
const addressVersion = 0x01

var (
	secp256k1_N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1_halfN = new(big.Int).Div(secp256k1_N, big.NewInt(2))
)

// ValidateSignatureValues verifies whether the signature values are valid with
// the given chain rules. The v value is assumed to be either 0 or 1.
func ValidateSignatureValues(v byte, r, s *big.Int) bool {
	if r.Cmp(math.Big1) < 0 || s.Cmp(math.Big1) < 0 {
		return false
	}
	// reject upper range of s values (ECDSA malleability)
	// see discussion in secp256k1/libsecp256k1/include/secp256k1.h
	if s.Cmp(secp256k1_halfN) > 0 {
		return false
	}
	// Allow r to be in full N range
	return r.Cmp(secp256k1_N) < 0 && s.Cmp(secp256k1_N) < 0 && (v == 0 || v == 1)
}

func PubToAddress(pub []byte) common.Address {
	return encodeAddress(pub[1:])
}

// encodeAddress encodes a data to lemo address
func encodeAddress(data []byte) common.Address {
	// Get the first 19 bits of the hash
	hashData := Keccak256(data)[:19]
	// Add version number
	versionHashData := append([]byte{addressVersion}, hashData...)
	// Type conversion
	return common.BytesToAddress(versionHashData)
}

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}
