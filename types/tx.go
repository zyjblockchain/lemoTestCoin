package types

import (
	"errors"
	"fmt"
	"github.com/lemoTestCoin/common"
	"math/big"
	"sync/atomic"
)

var (
	DefaultTTTL   uint64 = 2 * 60 * 60 // Transaction Time To Live, 2hours
	ErrInvalidSig        = errors.New("invalid transaction v, r, s values")
	TxVersion     uint8  = 1 // current transaction version. should between 0 and 128
)

type txdata struct {
	Recipient     *common.Address `json:"to" rlp:"nil"` // nil means contract creation
	RecipientName string          `json:"toName"`
	GasPrice      *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit      uint64          `json:"gasLimit" gencodec:"required"`
	Amount        *big.Int        `json:"amount" gencodec:"required"`
	Data          []byte          `json:"data"`
	Expiration    uint64          `json:"expirationTime" gencodec:"required"`
	Message       string          `json:"message"`

	// V is combined by these properties:
	//     type    version secp256k1.recovery  chainID
	// |----8----|----7----|--------1--------|----16----|
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}
type Transaction struct {
	data txdata

	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

// newTransaction
func newTransaction(version uint8, chainID uint16, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, expiration uint64, toName string, message string) *Transaction {
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

func (tx *Transaction) Type() uint8     { txType, _, _, _ := ParseV(tx.data.V); return txType }
func (tx *Transaction) Version() uint8  { _, version, _, _ := ParseV(tx.data.V); return version }
func (tx *Transaction) ChainID() uint16 { _, _, _, chainID := ParseV(tx.data.V); return chainID }

// WithSignature returns a new transaction with the given signature.
func (tx *Transaction) WithSignature(signer Signer, sig []byte) (*Transaction, error) {
	r, s, v, err := signer.ParseSignature(tx, sig)
	if err != nil {
		return nil, err
	}
	cpy := &Transaction{data: tx.data}
	cpy.data.R, cpy.data.S, cpy.data.V = r, s, v
	return cpy, nil
}

// CombineV combines type, version, chainID together to get V (without secp256k1.V)
func CombineV(txType uint8, version uint8, chainID uint16) *big.Int {
	return new(big.Int).SetUint64((uint64(txType) << 24) + (uint64(version&0x7f) << 17) + uint64(chainID))
}

// ParseV split V to 4 parts
func ParseV(V *big.Int) (txType uint8, version uint8, secp256k1V uint8, chainID uint16) {
	uint64V := V.Uint64()
	txType = uint8((uint64V >> 24) & 0xff)
	version = uint8((uint64V >> 17) & 0x7f)
	secp256k1V = uint8((uint64V >> 16) & 1)
	chainID = uint16(uint64V & 0xffff)
	return
}

// SetSecp256k1V merge secp256k1.V into the result of CombineV function
func SetSecp256k1V(V *big.Int, secp256k1V byte) *big.Int {
	// V = V & ((sig[64] & 1) << 16)
	return new(big.Int).SetBit(V, 16, uint(secp256k1V&1))
}
