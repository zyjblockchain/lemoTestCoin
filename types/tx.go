package types

import (
	"fmt"
	"math/big"
	"sync/atomic"
)

const (
	AddressLength = 20
	HashLength    = 32
)

type Address [AddressLength]byte

func StringToAddress(s string) (Address, error) {

}

type Hash [HashLength]byte

type txdata struct {
	Recipient     *Address `json:"to" rlp:"nil"` // nil means contract creation
	RecipientName string   `json:"toName"`
	GasPrice      *big.Int `json:"gasPrice" gencodec:"required"`
	GasLimit      uint64   `json:"gasLimit" gencodec:"required"`
	Amount        *big.Int `json:"amount" gencodec:"required"`
	Data          []byte   `json:"data"`
	Expiration    uint64   `json:"expirationTime" gencodec:"required"`
	Message       string   `json:"message"`

	// V is combined by these properties:
	//     type    version secp256k1.recovery  chainID
	// |----8----|----7----|--------1--------|----16----|
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *Hash `json:"hash" rlp:"-"`
}
type Transaction struct {
	data txdata

	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

// newTransaction
func newTransaction(version uint8, chainID uint16, to *Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, expiration uint64, toName string, message string) *Transaction {
	if version >= 128 {
		panic(fmt.Sprintf("invalid transaction version %d, should < 128", version))
	}
	d := txdata{
		Recipient:     to,
		RecipientName: toName,
		GasPrice:      new(big.Int),
		GasLimit:      gasLimit,
		Amount:        new(big.Int),
		Data:          data,
		Expiration:    expiration,
		Message:       message,
		V:             CombineV(0, version, chainID),
		R:             new(big.Int),
		S:             new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.GasPrice.Set(gasPrice)
	}
	return &Transaction{data: d}
}

// CombineV combines type, version, chainID together to get V (without secp256k1.V)
func CombineV(txType uint8, version uint8, chainID uint16) *big.Int {
	return new(big.Int).SetUint64((uint64(txType) << 24) + (uint64(version&0x7f) << 17) + uint64(chainID))
}
