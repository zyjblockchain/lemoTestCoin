package crypto

import "github.com/lemoTestCoin/common"

type AccountKey struct {
	Private string
	Public  string
	Address string
}

// GenerateAddress generate Lemo address
func GenerateAddress() (*AccountKey, error) {
	// Get privateKey
	privKey, err := GenerateKey()
	if err != nil {
		return nil, err
	}
	// Get the public key through the private key
	pubKey := privKey.PublicKey
	// Get the address(Address) through the public key
	address := PubkeyToAddress(pubKey)
	// get lemoAddress
	lemoAddress := address.String()

	// PublicKey type is converted to bytes type
	publicToBytes := FromECDSAPub(&pubKey)
	// PrivateKey type is converted to bytes type
	privateToBytes := FromECDSA(privKey)
	return &AccountKey{
		Private: common.ToHex(privateToBytes),
		Public:  common.ToHex(publicToBytes[1:]),
		Address: lemoAddress,
	}, nil
}
