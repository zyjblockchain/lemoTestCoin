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
	token          = "lemo"
	logo           = "Lemo"
	coinNum        = uint64(10000000000000000000) // 10 lemo
	interval       = uint64(24 * 3600)            // 间隔一天
	getBalanceFlag = "查询余额"                       // 用户发送查询余额请求的前缀标志位
)

// 判断request的类型
// TODO 目前只有区别扫码请求和普通文本的请求，如果将来要更加细分其他请求可以直接在此结构体下面加上字段来判断
type JudgeRequestType struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string
	FromUserName string
	MsgType      string
}

// 接收用户文本消息
type TextRequestBody struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string
	FromUserName string
	CreateTime   time.Duration
	MsgType      string
	Content      string
	MsgId        int
}

// 扫码事件响应数据结构
type EventRequestBody struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string
	FromUserName string
	CreateTime   time.Duration
	MsgType      string
	Event        string
	EventKey     string
	Ticket       string
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

// parseTextRequestBody 解析收到文本的消息
func parseTextRequestBody(body []byte) *TextRequestBody {
	// body, err := ioutil.ReadAll(r.Body)
	// defer r.Body.Close()
	// if err != nil {
	// 	log.Println("io read error", err)
	// 	return nil
	// }
	requestBody := &TextRequestBody{}
	err := xml.Unmarshal(body, requestBody)
	if err != nil {
		log.Println("xml.Unmarshal request error:", err)
	}
	// fmt.Println("TextRequestBody 结构体:", requestBody) // 调试用
	return requestBody
}

// 解析接收到的扫码事件的请求
func parseEventRequestBody(body []byte) *EventRequestBody {
	// body, err := ioutil.ReadAll(r.Body)
	// defer r.Body.Close()
	// if err != nil {
	// 	log.Println("io read error", err)
	// 	return nil
	// }
	requestBody := &EventRequestBody{}
	err := xml.Unmarshal(body, requestBody)
	if err != nil {
		log.Println("xml.Unmarshal request error:", err)
	}
	// fmt.Println("EventRequestBody 结构体:", requestBody) // 调试用
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
	// 检验url是否来自微信
	if !validateUrl(w, r) {
		log.Println("Wechat service: this http request is not from Wechat platform !")
		return
	}
	fmt.Println("请求方法:", r.Method) // 调试用
	// 微信端post请求
	if r.Method == "POST" {
		// 读取请求的body
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Println("io read error", err)
		}
		fmt.Println(string(body)) // 调试用

		judgeMsgType := &JudgeRequestType{}
		// 反序列化出body中的MsgType字段放在结构体 JudgeRequestType中
		err = xml.Unmarshal(body, judgeMsgType)
		if err != nil {
			log.Println("xml.Unmarshal error:", err)
		}
		fmt.Println("post请求的MsgType:", judgeMsgType.MsgType) // 调试用

		if judgeMsgType.MsgType == "event" { // 如果是事件请求则为扫码
			var err error
			var Msg string
			getRequestBody := parseEventRequestBody(body)     // 解析扫码事件请求
			fmt.Println("EventKey:", getRequestBody.EventKey) // 调试用
			// 判断新用户扫的二维码是否为我们水龙头推广的二维码
			if getRequestBody.EventKey == "" { // 满足此条件为微信公众号自带的二维码，说明只是用户扫此二维码关注公众号的操作
				Msg = "感谢关注LemoChain，请添加技术社区客服 Lucy180619 进入Lemo技术社区。"
			} else {
				// 新用户扫码关注推送消息如下
				if getRequestBody.Event == "subscribe" {
					Msg = "欢迎关注LemoChain,如需领取lemo测试币,请在公众号下方直接输入您的Lemo地址。"
				} else if getRequestBody.Event == "SCAN" { // 已关注的用户扫码进公众号
					Msg = "感谢使用Lemo测试币水龙头，请回复您的钱包地址，用于接收LEMO测试币。"
				}
			}
			responseTextBody, err = makeTextResponseBody(getRequestBody.ToUserName, getRequestBody.FromUserName, Msg)
			if err != nil {
				log.Println("Wechat Service: makeTextResponseBody error:", err)
				return
			}
		} else if judgeMsgType.MsgType == "text" { // 微信端发送文本消息的处理

			textRequestBody := parseTextRequestBody(body) // 解析文本请求
			// 判断用户发送过来的文本是Lemo地址
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
						fmt.Sprintf("请再次确认您的lemo地址\n%s \n\n10LEMO测试币将在1个工作日内发放至您的钱包。请添加技术社区客服 Lucy180619 进入Lemo技术社区。\n此次交易的哈希为%s\n", textRequestBody.Content, txHash))
					if err != nil {
						log.Println("Wechat Service: makeTextResponseBody error:", err)
						return
					}
				} else { // 不满足打币时间
					// 回复用户消息，为距离上次申请时间间隔小于24小时。
					responseTextBody, err = makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName,
						fmt.Sprintf("抱歉距离您上次申请时间小于24小时\n请在 %d小时 %d分钟 之后再次申请.", (interval-(uint64(time.Now().Unix()+30*60)-latestTime))/3600, ((interval-(uint64(time.Now().Unix()+30*60)-latestTime))%3600)/60))
					if err != nil {
						log.Println("Wechat Service: makeTextResponseBody error:", err)
						return
					}
				}
			} else if textRequestBody != nil && (textRequestBody.Content == "水龙头" || textRequestBody.Content == "测试币") {
				// 给用户提示申请打币的操作
				var err error
				responseTextBody, err = makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName,
					fmt.Sprint("感谢使用Lemo测试币水龙头,请回复您的钱包地址,用于接收LEMO测试币。"))
				if err != nil {
					log.Println("Wechat Service: makeTextResponseBody error:", err)
					return
				}
			} else if textRequestBody != nil && IsGetBalancePost(textRequestBody.Content) {
				// 解析用户地址
				lemoAdd := strings.TrimPrefix(textRequestBody.Content, getBalanceFlag)
				// 验证地址为正确的lemo地址
				if !fromLemoAddress(lemoAdd) {
					var err error
					responseTextBody, err = makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName,
						fmt.Sprint("输入的lemo地址不正确，请重新输入\n"))
					if err != nil {
						log.Println("Wechat Service: makeTextResponseBody error:", err)
						return
					}
				} else {
					// 进行查询操作
					balance, err := types.GetBalance(lemoAdd)
					if err != nil {
						responseTextBody, _ = makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName,
							fmt.Sprint("查询失败，请检查输入是否正确或者联系社区人员\n"))
					} else {
						responseTextBody, _ = makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName,
							balance)
					}
				}
			} else { // 用户发送未定义的消息给公众号的返回
				responseTextBody, _ = makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName,
					fmt.Sprint("输入数据请求无效。"))
			}

		} else { // 如果用户发送的消息是图片、视频、音频等类型，则不处理，以后有需求直接在此处扩展即可。
			responseTextBody, _ = makeTextResponseBody(judgeMsgType.ToUserName, judgeMsgType.FromUserName,
				fmt.Sprint("目前公众号还不支持此类消息,请谅解。"))
		}

		w.Header().Set("Content-Type", "text/xml")
		fmt.Println(string(responseTextBody))
		fmt.Fprintf(w, string(responseTextBody))
	}
}

// 验证content是一个可用的Lemo地址
func fromLemoAddress(content string) bool {
	if len(content) != 40 {
		return false
	}
	content = strings.ToUpper(content)
	return strings.HasPrefix(content, strings.ToUpper(logo))
}

// 验证用户发送的是一个查询账户余额的请求
func IsGetBalancePost(content string) bool {
	return strings.HasPrefix(content, getBalanceFlag)
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
// 	fmt.Println("start test!")
// 	err, txHash := types.SendCoin("Lemo83W7HDZYS33Z745NZ2FGF37565DSF5AHJZ4J", coinNum)
// 	if err != nil {
// 		fmt.Println("post err:", err)
// 	}
// 	fmt.Println("tx_Hash:", txHash)
// }
