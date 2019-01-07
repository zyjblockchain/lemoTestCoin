package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/lemoTestCoin/common"
	"github.com/lemoTestCoin/common/crypto"
	"github.com/lemoTestCoin/store"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"time"
)

const (
	defaultGasPrice = 1e9
	defaultGasLimit = 50000
	chainID         = 100
	chainUrl        = "http://149.28.68.93:8001" // 连接一个节点的ip地址
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
	wxTx := NewTransaction(to, new(big.Int).SetUint64(amount), defaultGasLimit, new(big.Int).SetUint64(defaultGasPrice), []byte{}, chainID, uint64(time.Now().Unix()+30*60), "wx", "water faucet")
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
	if err != nil {
		log.Println("marshal error:", err)
		return err, ""
	}
	fmt.Println(string(jsonData))
	reader := bytes.NewReader(jsonData)
	// post
	resp, err := http.Post(chainUrl, "application/json;charset=UTF-8", reader)
	if err != nil {
		log.Println("post error:", err)
		return err, ""
	}
	// 记录用户成功发起申请交易的时间到db
	err = store.Putdb(content, wxTx.data.Expiration)
	if err != nil {
		log.Println("put tx time to db error:", err)
		return err, ""
	}

	respTx, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println("glemo response error:", err)
		return err, ""
	}
	respon := &jsonSuccessResponse{}

	err = json.Unmarshal(respTx, respon)
	if err != nil {
		log.Println("json.Unmarshal response error:", err)
		return err, ""
	}
	return nil, respon.Result.(string)
}

type PostData struct {
	Version string            `json:"jsonrpc"`
	Id      uint64            `json:"id"`
	Method  string            `json:"method"`
	Payload []json.RawMessage `json:"params,omitempty"`
}
type jsonSuccessResponse struct {
	Version string      `json:"jsonrpc"`
	Id      uint64      `json:"id"`
	Result  interface{} `json:"result"`
}

// 查询用户账户余额
func GetBalance(lemoAddress string) (string, error) {

	jsonlemoAdd, err := json.Marshal(lemoAddress)
	if err != nil {
		log.Println("json 102 marshal error:", err)
	}
	data := PostData{
		Version: "2.0",
		Id:      1,
		Method:  "account_getBalance",
		Payload: []json.RawMessage{jsonlemoAdd},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("json 112 marshal error:", err)
		return "", err
	}
	reader := bytes.NewReader(jsonData)
	resp, err := http.Post(chainUrl, "application/json;charset=UTF-8", reader)
	if err != nil {
		log.Println("post error:", err)
		return "", err
	}
	byteResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("read response error:", err)
		return "", err
	}
	defer resp.Body.Close()
	respon := new(jsonSuccessResponse)
	err = json.Unmarshal(byteResp, respon)
	if err != nil {
		log.Println("unmarshal error:", err)
		return "", nil
	}
	return respon.Result.(string), nil
}
