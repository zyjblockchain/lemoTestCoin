package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/lemoTestCoin/common"
	"github.com/lemoTestCoin/common/crypto"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
)

const (
	defaultGasPrice = 1e9
	defaultGasLimit = 50000
	chainID         = 1
)

// senderLemoAddress = "Lemo83GN72GYH2NZ8BA729Z9TCT7KQ5FC3CR6DJG"
var SenderToPrivate, _ = crypto.HexToECDSA("c21b6b2fbf230f665b936194d14da67187732bf9d28768aef1a3cbb26608f8aa")

// 与glemo交互发送交易
func SendCoin(content string, amount uint64) (error, string) {
	to, err := common.StringToAddress(content)
	if err != nil {
		log.Println("decode address error:", err)
		return err, ""
	}
	// 生成交易
	wxTx := NewTransaction(to, new(big.Int).SetUint64(amount), defaultGasLimit, new(big.Int).SetUint64(defaultGasPrice), []byte{}, chainID, 0, "wx", "water faucet")
	// 签名交易
	signWxTx := SignTransaction(wxTx, SenderToPrivate)
	txData, err := signWxTx.MarshalJSON()
	fmt.Println(string(txData))
	if err != nil {
		log.Println("tx marshal failed:", err)
		return err, ""
	}
	// jsonData := []byte(`{"jsonrpc": "2.0","method": "tx_sendTx","params": [],"id": 1}`)
	data := &PostData{
		Version: "2.0",
		Id:      1,
		Method:  "tx_sendTx",
		Payload: []json.RawMessage{txData},
	}
	jsonData, err := json.Marshal(data)
	fmt.Println(string(jsonData))
	reader := bytes.NewReader(jsonData)
	// post
	resp, err := http.Post("http://127.0.0.1:8001", "application/json;charset=UTF-8", reader)
	if err != nil {
		log.Println("post error:", err)
		return err, ""
	}
	respTx, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println("glemo response error:", err)
		return err, ""
	}
	return nil, string(respTx)
}

type PostData struct {
	Version string            `json:"jsonrpc"`
	Id      uint64            `json:"id"`
	Method  string            `json:"method"`
	Payload []json.RawMessage `json:"params,omitempty"`
}
