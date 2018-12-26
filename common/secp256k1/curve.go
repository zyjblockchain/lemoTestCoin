package secp256k1

import (
	"github.com/lemoTestCoin/common/math"
	"math/big"
)

type BitCurve struct {
	P       *big.Int // the order of the underlying field
	N       *big.Int // the order of the base point
	B       *big.Int // the constant of the BitCurve equation
	Gx, Gy  *big.Int // (x,y) of the base point
	BitSize int      // the size of the underlying field
}

// Marshal converts a point into the form specified in section 4.3.6 of ANSI
// X9.62.
func (BitCurve *BitCurve) Marshal(x, y *big.Int) []byte {
	byteLen := (BitCurve.BitSize + 7) >> 3
	ret := make([]byte, 1+2*byteLen)
	ret[0] = 4 // uncompressed point flag
	math.ReadBits(x, ret[1:1+byteLen])
	math.ReadBits(y, ret[1+byteLen:])
	return ret
}

var theCurve = new(BitCurve)

func init() {
	// See SEC 2 section 2.7.1
	// curve parameters taken from:
	// http://www.secg.org/collateral/sec2_final.pdf
	theCurve.P, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	theCurve.N, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
	theCurve.B, _ = new(big.Int).SetString("0000000000000000000000000000000000000000000000000000000000000007", 16)
	theCurve.Gx, _ = new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	theCurve.Gy, _ = new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)
	theCurve.BitSize = 256
}

// S256 returns a BitCurve which implements secp256k1.
func S256() *BitCurve {
	return theCurve
}
