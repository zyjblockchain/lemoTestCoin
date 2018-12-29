package main

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"github.com/lemoTestCoin/common/store"
	"github.com/lemoTestCoin/types"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	token    = "lemo"
	logo     = "Lemo"
	coinNum  = uint64(10000000000000000000) // 10 lemo
	interval = uint64(24 * 3600)            // 间隔一天
)

// 接收用户消息
type TextRequestBody struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string
	FromUserName string
	CreateTime   time.Duration
	MagType      string
	Content      string
	MsgId        int
}

// 响应用户的消息
type TextResponseBody struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   CDATAText
	FromUserName CDATAText
	CreateTime   time.Duration
	MsgType      CDATAText
	Content      CDATAText
}

// CDATAText
type CDATAText struct {
	Text string `xml:",innerxml"`
}

// makeSignature 生成签名
func makeSignature(timestamp, nonce string) string {
	s1 := []string{token, timestamp, nonce}
	sort.Strings(s1)
	s := sha1.New()
	io.WriteString(s, strings.Join(s1, ""))
	return fmt.Sprintf("%x", s.Sum(nil))
}

// validateUrl 检验url是否来自微信
func validateUrl(w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	timestamp := strings.Join(r.Form["timestamp"], "")
	nonce := strings.Join(r.Form["nonce"], "")
	signatureGen := makeSignature(timestamp, nonce)

	signatureIn := strings.Join(r.Form["signature"], "")
	if signatureGen != signatureIn {
		return false
	}
	echostr := strings.Join(r.Form["echostr"], "")
	fmt.Fprintf(w, echostr)
	return true
}

// parseTextRequestBody 解析收到的消息
func parseTextRequestBody(r *http.Request) *TextRequestBody {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
		return nil
	}
	fmt.Println(string(body))
	requestBody := &TextRequestBody{}
	err = xml.Unmarshal(body, requestBody)
	if err != nil {
		log.Println("xml.Unmarshal request error:", err)
	}
	return requestBody
}

// value2CDATA
func value2CDATA(v string) CDATAText {
	return CDATAText{"<![CDATA[" + v + "]]>"}
}

// 生成server响应的xml
func makeTextResponseBody(fromUserName, toUserName, content string) ([]byte, error) {
	textResponseBody := &TextResponseBody{}
	textResponseBody.FromUserName = value2CDATA(fromUserName)
	textResponseBody.ToUserName = value2CDATA(toUserName)
	textResponseBody.MsgType = value2CDATA("text")
	textResponseBody.Content = value2CDATA(content)
	textResponseBody.CreateTime = time.Duration(time.Now().Unix())
	return xml.MarshalIndent(textResponseBody, " ", "  ")
}

// server服务
func procRequest(w http.ResponseWriter, r *http.Request) {
	var responseTextBody []byte
	if !validateUrl(w, r) {
		log.Println("Wechat service: this http request is not from Wechat platform !")
		return
	}
	if r.Method == "POST" {
		textRequestBody := parseTextRequestBody(r)
		// 判断用户发送过来的content是Lemo地址
		if textRequestBody != nil && fromLemoAddress(textRequestBody.Content) {
			// fmt.Printf("Wechat service: Received text msg [%s] from user [%s]!\n",
			// 	textRequestBody.Content, textRequestBody.FromUserName)

			// 获取用户上次申请打币的时间
			latestTime, err := store.Getdb(textRequestBody.Content)
			if err != nil {
				log.Println("get db error:", err)
				return
			}
			// 满足打币的条件:
			// 1. 在24小时之后才能打币
			// 2. (latestTime == 0)表示db里没有此用户记录,用户为第一次申请打币
			if latestTime+interval < uint64(time.Now().Unix()+30*60) || latestTime == 0 {
				// 获取到用户的地址进行打币操作...
				err, txHash := types.SendCoin(textRequestBody.Content, coinNum)
				if err != nil {
					log.Println("send coin error:", err)
					return
				}
				// 打印出glemo返回的交易信息
				// fmt.Println(txHash)
				// 回复用户消息
				responseTextBody, err = makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName,
					fmt.Sprintf("请再次确认您的lemo地址\n {%s} \n\n 10LEMO测试币将在1个工作日内发放至您的钱包.\n 请添加技术社区客服 Lucy180619 进入Lemo技术社区.\r\n 此次交易的哈希为{%s}\n", textRequestBody.Content, txHash))
				if err != nil {
					log.Println("Wechat Service: makeTextResponseBody error:", err)
					return
				}
			} else { // 不满足打币时间
				// 回复用户消息，为距离上次申请时间间隔小于24小时。
				responseTextBody, err = makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName,
					fmt.Sprintf("抱歉距离您上次申请时间小于24小时\n 请在 %s小时 %s分钟 之后再次申请.", (interval-(uint64(time.Now().Unix()+30*60)-latestTime))/3600, ((interval-(uint64(time.Now().Unix()+30*60)-latestTime))%3600)/60+1))
				if err != nil {
					log.Println("Wechat Service: makeTextResponseBody error:", err)
					return
				}
			}
			w.Header().Set("Content-Type", "text/xml")
			fmt.Println(string(responseTextBody))
			fmt.Fprintf(w, string(responseTextBody))
		}
	}
}

// 验证content是一个可用的Lemo地址
func fromLemoAddress(content string) bool {
	content = strings.ToUpper(content)
	return strings.HasPrefix(content, strings.ToUpper(logo))
}

func main() {
	log.Println("Wechat Service: Start!")
	http.HandleFunc("/", procRequest)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal("Wechat Service: ListenAndServer failed,", err)
	}
	log.Println("Wechat Service: Stop!")
}

// // 测试用
// func main1() {
// 	fmt.Println("start test!")
// 	err, txHash := types.SendCoin("Lemo83N65NKDY8D2FKKQY35JWJTCPZ8DSYHP7GPT", 10000)
// 	if err != nil {
// 		fmt.Println("post err:", err)
// 	}
// 	fmt.Println("tx_Hash:", txHash)
// }
