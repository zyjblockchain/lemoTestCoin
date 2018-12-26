package common

import (
	"errors"
	"github.com/lemoTestCoin/common/base26"

	"strings"
)

const (
	AddressLength = 20
	HashLength    = 32
	logo          = "Lemo"
)

type Address [AddressLength]byte

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// Sets the address to the value of b. If b is larger than len(a) it will panic
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
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

func StringToAddress(s string) (Address, error) {
	if isLemoAddress(s) {
		var a Address
		err := a.Decode(s)
		return a, err
	}
	return BytesToAddress([]byte(s)), nil
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

type Hash [HashLength]byte
