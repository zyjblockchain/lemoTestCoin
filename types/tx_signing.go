package types

import (
	"crypto/ecdsa"
	"errors"
	"github.com/lemoTestCoin/common/crypto"
	"github.com/lemoTestCoin/common/crypto/sha3"
	"github.com/lemoTestCoin/common/rlp"

	"github.com/lemoTestCoin/common"
)

var (
	ErrPublicKey   = errors.New("invalid public key")
	ErrNoSignsData = errors.New("no signature data")
)

// MakeSigner returns a Signer based on the given version and chainID.
func MakeSigner() Signer {
	return DefaultSigner{}
}

// recoverSigners
func recoverSigners(sigHash common.Hash, sigs [][]byte) ([]common.Address, error) {
	length := len(sigs)
	if length == 0 {
		return nil, ErrNoSignsData
	}
	signers := make([]common.Address, length, length)
	for i := 0; i < length; i++ {
		// recover the public key from the signature
		pub, err := crypto.Ecrecover(sigHash[:], sigs[i])
		if err != nil {
			return nil, err
		}
		if len(pub) == 0 || pub[0] != 4 {
			return nil, ErrPublicKey
		}
		addr := crypto.PubToAddress(pub)
		signers[i] = addr
	}
	return signers, nil
}

// Signer encapsulates transaction signature handling.
type Signer interface {
	// SignTx returns transaction after signature
	SignTx(tx *Transaction, prv *ecdsa.PrivateKey) (*Transaction, error)

	// GetSigners returns the sender address of the transaction.
	GetSigners(tx *Transaction) ([]common.Address, error)

	// Hash returns the hash to be signed.
	Hash(tx *Transaction) common.Hash
}

// DefaultSigner implements Signer.
type DefaultSigner struct {
}

func (s DefaultSigner) SignTx(tx *Transaction, prv *ecdsa.PrivateKey) (*Transaction, error) {
	h := s.Hash(tx)

	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	cpy := tx.Clone()
	cpy.data.Sigs = append(cpy.data.Sigs, sig)
	return cpy, nil
}

func (s DefaultSigner) GetSigners(tx *Transaction) ([]common.Address, error) {
	sigHash := s.Hash(tx)

	sigs := tx.data.Sigs
	signers, err := recoverSigners(sigHash, sigs)
	return signers, err
}
func (s DefaultSigner) Hash(tx *Transaction) common.Hash {
	hashData := getHashData(tx)

	return rlpHash([]interface{}{
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
	})
}

// rlpHash 数据rlp编码后求hash
func rlpHash(data interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, data)
	hw.Sum(h[:0])
	return h
}
