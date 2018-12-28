package main

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
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
	token = "lemo"
	logo  = "Lemo"
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
	if err != nil {
		log.Fatal(err)
		return nil
	}
	fmt.Println(string(body))
	requestBody := &TextRequestBody{}
	xml.Unmarshal(body, requestBody)
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
			// 获取到用户的地址进行打币操作，包括判断是否在24小时之内打过币...
			err, txMsg := types.SendCoin(textRequestBody.Content, 10000)
			if err != nil {
				log.Println("send coin error:", err)
				return
			}
			// 打印出glemo返回的交易信息
			fmt.Println(txMsg)
			// 回复用户消息
			responseTextBody, err := makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName,
				fmt.Sprintf("Your Lemo Address [%s] will get [10000] lemo;\n please wait for some time;\n", textRequestBody.Content)+txMsg)
			if err != nil {
				log.Println("Wechat Service: makeTextResponseBody error:", err)
				return
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
// func main() {
// 	err, txMsg := types.SendCoin("Lemo83J686JCF3PQDKN6AH6QF4W3P9NK2JQY9G2W", 10000)
// 	if err != nil {
// 		fmt.Println("post err:", err)
// 	}
// 	fmt.Println(txMsg)
// }
