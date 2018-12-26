package types

import (
	"log"
	"net/rpc"
)

// 与glemo交互发送交易
func sendTx(to string, amount uint64) bool {
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:8001")
	if err != nil {
		log.Fatal("dialing:", err)
		return false
	}

	tx := newTransaction(1, 100)
	client.Call("NewPublicTxAPI.SendTx", tx)

}
