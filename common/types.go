package common

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/lemoTestCoin/common/base26"
	"github.com/lemoTestCoin/common/hexutil"
	"math/big"
	"math/rand"
	"reflect"

	"strings"
)

const (
	AddressLength = 20
	HashLength    = 32
	logo          = "Lemo"
)

var (
	hashT    = reflect.TypeOf(Hash{})
	addressT = reflect.TypeOf(Address{})
)

type Hash [HashLength]byte

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}
func StringToHash(s string) Hash { return BytesToHash([]byte(s)) }
func BigToHash(b *big.Int) Hash  { return BytesToHash(b.Bytes()) }
func HexToHash(s string) Hash    { return BytesToHash(FromHex(s)) }

// Get the string representation of the underlying hash
func (h Hash) Str() string   { return string(h[:]) }
func (h Hash) Bytes() []byte { return h[:] }
func (h Hash) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }
func (h Hash) Hex() string   { return hexutil.Encode(h[:]) }

// TerminalString implements log.TerminalStringer, formatting a string for console
// output during logging.
func (h Hash) TerminalString() string {
	return fmt.Sprintf("%x…%x", h[:3], h[29:])
}

// String implements the stringer interface and is used also by the logger when
// doing full logging into a file.
func (h Hash) String() string {
	return h.Hex()
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (h Hash) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "%"+string(c), h[:])
}

// UnmarshalText parses a hash in hex syntax.
func (h *Hash) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Hash", input, h[:], true)
}

// UnmarshalJSON parses a hash in hex syntax.
func (h *Hash) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(hashT, input, h[:])
}

// MarshalText returns the hex representation of h.
func (h Hash) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// Sets the hash to the value of b. If b is larger than len(h), 'b' will be cropped (from the left).
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

// Set string `s` to h. If s is larger than len(h) s will be cropped (from left) to fit.
func (h *Hash) SetString(s string) { h.SetBytes([]byte(s)) }

// Sets h to other
func (h *Hash) Set(other Hash) {
	for i, v := range other {
		h[i] = v
	}
}

// Generate implements testing/quick.Generator.
func (h Hash) Generate(rand *rand.Rand, size int) reflect.Value {
	m := rand.Intn(len(h))
	for i := len(h) - 1; i > m; i-- {
		h[i] = byte(rand.Uint32())
	}
	return reflect.ValueOf(h)
}

func EmptyHash(h Hash) bool {
	return h == Hash{}
}

// Address

// Address represents the 20 byte address of an Lemochain account.
type Address [AddressLength]byte

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}
func StringToAddress(s string) (Address, error) {
	if isLemoAddress(s) {
		var a Address
		err := a.Decode(s)
		return a, err
	}
	return BytesToAddress([]byte(s)), nil
}
func BigToAddress(b *big.Int) Address { return BytesToAddress(b.Bytes()) }
func HexToAddress(s string) Address   { return BytesToAddress(FromHex(s)) }

// IsHexAddress verifies whether a string can represent a valid hex-encoded
// Lemochain address or not.
func IsHexAddress(s string) bool {
	if hasHexPrefix(s) {
		s = s[2:]
	}
	return len(s) == 2*AddressLength && isHex(s)
}

// Get the string representation of the underlying address
func (a Address) Str() string   { return string(a[:]) }
func (a Address) Bytes() []byte { return a[:] }
func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }
func (a Address) Hash() Hash    { return BytesToHash(a[:]) }

// Hex returns an EIP55-compliant hex string representation of the address.
func (a Address) Hex() string {
	address := hex.EncodeToString(a[:])
	return "0x" + string(address)
}

// String implements the stringer interface and native address is converted to LemoAddress.
func (a Address) String() string {
	// Get check digit
	checkSum := GetCheckSum(a.Bytes())
	// Stitching the check digit at the end
	fullPayload := append(a.Bytes(), checkSum)
	// base26 encoding
	bytesAddress := base26.Encode(fullPayload)
	// Add logo at the top
	lemoAddress := strings.Join([]string{logo, bytesAddress}, "")

	return lemoAddress
}

// Decode decodes original address by the LemoAddress format.
func (a *Address) Decode(lemoAddress string) error {
	if !isLemoAddress(lemoAddress) {
		return errors.New("address decode fail")
	}
	lemoAddress = strings.ToUpper(lemoAddress)
	// Remove logo
	address := []byte(lemoAddress)[len(logo):]
	// Base26 decoding
	fullPayload := base26.Decode(address)
	// get the length of the address bytes type
	length := len(fullPayload)
	if length == 0 {
		// 0x0000000000000000000000000000000000000000
		a.SetBytes(nil)
	} else {
		// get check bit
		checkSum := fullPayload[length-1]
		// get the native address
		bytesAddress := fullPayload[:length-1]
		// calculate the check bit by bytesAddress
		trueCheck := GetCheckSum(bytesAddress)
		// compare check
		if checkSum != trueCheck {
			return errors.New("lemo address check fail")
		}
		a.SetBytes(bytesAddress)
	}
	return nil
}

func isLemoAddress(str string) bool {
	str = strings.ToUpper(str)
	return strings.HasPrefix(str, strings.ToUpper(logo))
}

// GetCheckSum get the check digit by doing an exclusive OR operation
func GetCheckSum(addressToBytes []byte) byte {
	var temp = byte(0)
	for _, c := range addressToBytes {
		temp ^= c
	}
	return temp
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (a Address) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "%"+string(c), a[:])
}

// Sets the address to the value of b. If b is larger than len(a) it will panic
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// Set string `s` to a. If s is larger than len(a) it will panic
func (a *Address) SetString(s string) { a.SetBytes([]byte(s)) }

// Sets a to other
func (a *Address) Set(other Address) {
	for i, v := range other {
		a[i] = v
	}
}

// MarshalText returns the hex representation of a.
func (a Address) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalText parses a hash in hex syntax.
func (a *Address) UnmarshalText(input []byte) error {
	if isLemoAddress(string(input)) {
		return a.Decode(string(input))
	}
	return hexutil.UnmarshalFixedText("Address", input, a[:], true)
}

// UnmarshalJSON parses a hash in hex syntax.
func (a *Address) UnmarshalJSON(input []byte) error {
	originUpper := strings.Trim(strings.ToUpper(string(input)), "\"")
	if isLemoAddress(originUpper) {
		return a.Decode(originUpper)
	}
	return hexutil.UnmarshalFixedJSON(addressT, input, a[:])
}

type AddressSlice []Address

func (a AddressSlice) Len() int {
	return len(a)
}

func (a AddressSlice) Less(i, j int) bool {
	return a[i].Hex() < a[j].Hex()
}

func (a AddressSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
