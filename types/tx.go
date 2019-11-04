package types

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-core/chain/params"
	"github.com/lemoTestCoin/common"
	"github.com/lemoTestCoin/common/hexutil"
	"github.com/lemoTestCoin/common/rlp"
	"io"
	"math/big"
	"sync/atomic"
)

//go:generate gencodec -type txdata --field-override txdataMarshaling -out gen_tx_json.go

var (
	DefaultTTTL   uint64 = 2 * 60 * 60 // Transaction Time To Live, 2hours
	ErrInvalidSig        = errors.New("invalid transaction v, r, s values")
	TxVersion     uint8  = 1 // current transaction version. should between 0 and 128
)

type Transactions []*Transaction

type Transaction struct {
	data txdata
	hash atomic.Value
	size atomic.Value
}

type txdata struct {
	Type          uint16          `json:"type" gencodec:"required"`
	Version       uint8           `json:"version" gencodec:"required"`
	ChainID       uint16          `json:"chainID" gencodec:"required"`
	From          common.Address  `json:"from" gencodec:"required"`
	GasPayer      *common.Address `json:"gasPayer" rlp:"nil"`
	Recipient     *common.Address `json:"to" rlp:"nil"` // nil means contract creation
	RecipientName string          `json:"toName"`
	GasPrice      *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit      uint64          `json:"gasLimit" gencodec:"required"`
	GasUsed       uint64          `json:"gasUsed"`
	Amount        *big.Int        `json:"amount" gencodec:"required"`
	Data          []byte          `json:"data"`
	Expiration    uint64          `json:"expirationTime" gencodec:"required"`
	Message       string          `json:"message"`
	Sigs          [][]byte        `json:"sigs" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
	// gas payer signature
	GasPayerSigs [][]byte `json:"gasPayerSigs"`
}

type txdataMarshaling struct {
	Type         hexutil.Uint16
	Version      hexutil.Uint8
	ChainID      hexutil.Uint16
	GasPrice     *hexutil.Big10
	GasLimit     hexutil.Uint64
	GasUsed      hexutil.Uint64
	Amount       *hexutil.Big10
	Data         hexutil.Bytes
	Expiration   hexutil.Uint64
	Sigs         []hexutil.Bytes
	GasPayerSigs []hexutil.Bytes
}

// Clone deep copy transaction
func (tx *Transaction) Clone() *Transaction {
	cpy := *tx
	// Clear old hash. So we can change any field in the new tx. It will be created again when Hash() is called
	cpy.hash = atomic.Value{}

	if tx.data.Recipient != nil {
		*cpy.data.Recipient = *tx.data.Recipient
	}
	*cpy.data.GasPayer = *tx.data.GasPayer

	if tx.data.Sigs != nil {
		cpy.data.Sigs = make([][]byte, len(tx.data.Sigs), len(tx.data.Sigs))
		copy(cpy.data.Sigs, tx.data.Sigs)
	}
	if tx.data.GasPayerSigs != nil {
		cpy.data.GasPayerSigs = make([][]byte, len(tx.data.GasPayerSigs), len(tx.data.GasPayerSigs))
		copy(cpy.data.GasPayerSigs, tx.data.GasPayerSigs)
	}
	if tx.data.Data != nil {
		cpy.data.Data = make([]byte, len(tx.data.Data), len(tx.data.Data))
		copy(cpy.data.Data, tx.data.Data)
	}
	if tx.data.Hash != nil {
		*cpy.data.Hash = *tx.data.Hash
	}
	if tx.data.GasPrice != nil {
		cpy.data.GasPrice = new(big.Int).Set(tx.data.GasPrice)
	}
	if tx.data.Amount != nil {
		cpy.data.Amount = new(big.Int).Set(tx.data.Amount)
	}
	return &cpy
}

func NewTransaction(from common.Address, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, TxType uint16, chainID uint16, expiration uint64, toName string, message string) *Transaction {
	return newTransaction(from, TxType, TxVersion, chainID, nil, &to, amount, gasLimit, gasPrice, data, expiration, toName, message)
}

// SignTransaction 对交易签名
func SignTransaction(tx *Transaction, private *ecdsa.PrivateKey) *Transaction {
	signer := MakeSigner()
	tx, err := signer.SignTx(tx, private)
	if err != nil {
		panic(err)
	}
	return tx
}

// newTransaction
func newTransaction(from common.Address, txType uint16, version uint8, chainID uint16, gasPayer, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, expiration uint64, toName string, message string) *Transaction {
	if version >= 128 {
		panic(fmt.Sprintf("invalid transaction version %d, should < 128", version))
	}
	if gasPayer == nil {
		gasPayer = &from
	}
	d := txdata{
		Type:          txType,
		Version:       version,
		ChainID:       chainID,
		From:          from,
		GasPayer:      gasPayer,
		Recipient:     to,
		RecipientName: toName,
		GasPrice:      new(big.Int),
		GasLimit:      gasLimit,
		Amount:        new(big.Int),
		Data:          data,
		Expiration:    expiration,
		Message:       message,
		Sigs:          make([][]byte, 0),
		Hash:          nil,
		GasPayerSigs:  make([][]byte, 0),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.GasPrice.Set(gasPrice)
	}
	return &Transaction{data: d}
}

// EncodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &tx.data)
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}

	return err
}
func (tx *Transaction) Type() uint16    { return tx.data.Type }
func (tx *Transaction) Version() uint8  { return tx.data.Version }
func (tx *Transaction) ChainID() uint16 { return tx.data.ChainID }

// 箱子交易
//go:generate gencodec -type Box -out gen_box_json.go
type Box struct {
	SubTxList Transactions `json:"subTxList"  gencodec:"required"`
}

// GetBox
func GetBox(txData []byte) (*Box, error) {
	box := &Box{}
	err := json.Unmarshal(txData, box)
	if err != nil {
		return nil, err
	}
	return box, nil
}

// getHashData 获取计算交易hash 的交易data
func getHashData(tx *Transaction) interface{} {
	if tx.Type() == params.BoxTx {
		box, err := GetBox(tx.data.Data)
		if err != nil {
			// 箱子交易反序列化失败，把它当做普通交易处理
			return tx.data.Data
		}
		// 箱子中存在子交易
		if len(box.SubTxList) > 0 {
			return calcBoxSubTxHashSet(box.SubTxList)
		}
	}
	return tx.data.Data
}

// calcBoxSubTxHashSet 计算子交易的hash集合,返回对集合的hash值
func calcBoxSubTxHashSet(subTxList Transactions) common.Hash {
	// 计算子交易的交易hash集合
	subTxHashSet := make([]common.Hash, 0, len(subTxList))
	for _, subTx := range subTxList {
		subTxHashSet = append(subTxHashSet, subTx.Hash())
	}
	// 对子交易的交易hash集合进行hash 作为计算box交易hash 的data
	return rlpHash(subTxHashSet)
}

// Hash
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}

	hashData := getHashData(tx)
	result := rlpHash([]interface{}{
		tx.Type(),
		tx.Version(),
		tx.ChainID(),
		tx.data.From,
		tx.data.GasPayer,
		tx.data.Recipient,
		tx.data.RecipientName,
		tx.data.GasPrice,
		tx.data.GasLimit,
		tx.data.Amount,
		hashData,
		tx.data.Expiration,
		tx.data.Message,
		tx.data.Sigs,
		tx.data.GasPayerSigs,
	})

	tx.hash.Store(result)
	return result
}

// MarshalJSON encodes the lemoClient RPC transaction format.
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	hash := tx.Hash()
	data := tx.data
	data.Hash = &hash
	return data.MarshalJSON()
}

// CombineV combines type, version, chainID together to get V (without secp256k1.V)
func CombineV(txType uint8, version uint8, chainID uint16) *big.Int {
	return new(big.Int).SetUint64((uint64(txType) << 24) + (uint64(version&0x7f) << 17) + uint64(chainID))
}

// SetSecp256k1V merge secp256k1.V into the result of CombineV function
func SetSecp256k1V(V *big.Int, secp256k1V byte) *big.Int {
	// V = V & ((sig[64] & 1) << 16)
	return new(big.Int).SetBit(V, 16, uint(secp256k1V&1))
}
