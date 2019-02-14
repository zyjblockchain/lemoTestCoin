package main

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"github.com/lemoTestCoin/common/crypto"
	"github.com/lemoTestCoin/manager"
	"github.com/lemoTestCoin/store"
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
	token          = "lemo" // 用于验证来自绑定的微信公众号的请求的token，与公众号中的设置相同
	logo           = "Lemo"
	coinNum        = uint64(10000000000000000000)       // 10 lemo
	interval       = uint64(24 * 3600)                  // 每个lemo地址限制申请测试币的间隔时间为一天
	getBalanceFlag = "余额"                               // 用户发送查询余额请求的前缀标志位
	AppID          = "wx06904bf57c491375"               // 绑定公众号的appid ,只有绑定此公众号上发送过来的用户才能被标记。其他链接此水龙头的公众号上的用户能申请到测试币但是不能被标记。
	AppSecret      = "b8a85d73ac52df0e5552fb25793382b5" // 绑定公众号的app秘钥
	TagName        = "开发者"                              // 标记申请测试币的微信用户的组名
	// AppID     = "wx1f6db98d761e4679"
	// AppSecret = "fca7d817f6706c240cfbef7d554db891"
)

// 存储access_token 的全局变量
var AccessToken string

// 存储标签的全局变量
var tagId int

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
				Msg = "欢迎来到LemoChain中文官方社区大本营，点击lemochain.com发现更多关于LemoChain的信息！\n\n1）如果你想知道什么是LemoChain，请点击这里：https://mp.weixin.qq.com/s/9vJ4n7JkVExkolMu1AhDnA\n\n2）我们肯定是整个币圈，最有逼格的团队：https://mp.weixin.qq.com/s/eTjh9MB60VbMLt14mqlYSw\n\n3）说了这么多，LemoChain到底有什么用？https://mp.weixin.qq.com/s/WZcPL__zap14ryR9G3uwZQ\n\n4）既然都这么了解我们了，要不要点击下方【加入社区】成为LemoChain的一员呢？\n\n5）Lemo测试网已上线，回复'水龙头'获取测试网Lemo。"
			} else {
				// 新用户扫码关注推送消息如下
				if getRequestBody.Event == "subscribe" {
					Msg = "欢迎关注LemoChain,如需领取测试网Lemo,请在公众号下方直接输入您的Lemo地址。"
				} else if getRequestBody.Event == "SCAN" { // 已关注的用户扫码进公众号
					Msg = "感谢使用Lemo水龙头，请回复您的钱包地址，用于接收测试网Lemo。"
				} else {
					Msg = "欢迎进入LemoChain社区" // 防止给微信的响应为nil从而报错
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

					// 标记用户标签为 “开发者”,为了不能重复标记，通过判断用户最新打币时间来判断，当最新打币时间 latestTime==0 则此用户第一次申请打币则标记
					if latestTime == 0 {
						err = manager.AddTagForUser(AccessToken, []string{textRequestBody.FromUserName}, tagId)
						if err != nil {
							log.Println("给用户标记标签error:", err)
							return
						}
					}
					// 回复用户消息
					responseTextBody, err = makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName,
						fmt.Sprintf("请再次确认您的lemo地址\n%s \n\n测试网10LEMO将在1个工作日内发放至您的钱包。\n请添加技术社区客服微信 Lucy180619 进入Lemo技术社区。\n\n此次交易的哈希为%s\n", textRequestBody.Content, txHash))
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
							fmt.Sprint("查询失败，请检查输入是否正确或者联系技术社区客服微信 Lucy180619\n"))
					} else {
						responseTextBody, _ = makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName,
							balance)
					}
				}
			} else if textRequestBody != nil { // 用户在公众号输入'测试币','水龙头','LemoChain','官网','交易','周报',这些词汇跳转到的回复逻辑如下
				var respMsg string
				var err error
				if textRequestBody.Content == "水龙头" || textRequestBody.Content == "测试币" {
					respMsg = "感谢使用Lemo水龙头,请回复您的Lemo地址,用于接收测试网LEMO。\n若无Lemo地址，请回复'申请测试账户'获取Lemo地址。"
				} else if textRequestBody.Content == "LemoChain" || textRequestBody.Content == "lemochain" || textRequestBody.Content == "Lemochain" || textRequestBody.Content == "lemoChain" {
					respMsg = "LemoChain是一个非盈利社区化的区块链项目，其团队由来自硅谷、新加坡、伦敦、成都等高科技人士组成，为不同行业的应用开发者和服务商提供去中心化的用户账户系统、数据流通服务、 数字资产确权及用户诚信协议，构建未来应用的数字资产生态体系。"
				} else if textRequestBody.Content == "官网" {
					respMsg = "藏的那么深还是被你发现了，点击https://www.lemochain.com进入LemoChain更加深入的了解Lemo吧！"
				} else if textRequestBody.Content == "交易" {
					respMsg = "LEMO现已上线Gate交易所，点击https://www.gate.io/即可查看哦"
				} else if textRequestBody.Content == "周报" {
					respMsg = "谢谢你对我如此关心，点击下方【历史消息】菜单按钮就可以查看往期周报了哦。"
				} else if textRequestBody.Content == "申请测试账户" {
					accountKey, _ := crypto.GenerateAddress()
					LemoAddr := accountKey.Address
					// PubKey := accountKey.Public
					Private := accountKey.Private
					respMsg = fmt.Sprintf("该账户仅供测试网使用，请妥善保存您的地址及私钥。\n\n复制并回复您以Lemo开头的地址，即可获取测试网LEMO，测试网LEMO仅供测试网使用。\n\n私钥：\n%s\n地址：\n%s", Private, LemoAddr)
				} else if textRequestBody.Content == "214" || textRequestBody.Content == "情人节" {
					respMsg = "感谢参与Lemo情人节活动！\n\n点击下面链接，填写领奖信息\n\n海量LEMO等你拿！\n\nhttps://dwz.cn/UvOjE9Ci"
				} else { // 用户发送的是未定义的text内容
					respMsg = "感谢关注LemoChain，点击右下角【加入社群】菜单按钮，和柠檬粉们一起嗨～"
				}
				responseTextBody, err = makeTextResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName, respMsg)
				if err != nil {
					log.Println("Wechat Service: makeTextResponseBody error:", err)
					return
				}
			}

		} else { // 如果用户发送的消息是图片、视频、音频等类型，则不处理，以后有需求直接在此处扩展即可。
			responseTextBody, _ = makeTextResponseBody(judgeMsgType.ToUserName, judgeMsgType.FromUserName,
				fmt.Sprint("感谢关注LemoChain，点击右下角【加入社群】菜单按钮，和柠檬粉们一起嗨～"))
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

	// --------------------获取access_token----------------------------- //
	// 获取access_token
	var err error
	AccessToken, err = manager.GetAccessToken(AppID, AppSecret)
	fmt.Println("启动时生成的access_token:", AccessToken)
	if err != nil {
		log.Fatal("get access_token error:", err)

	}
	// 定时更新access_token
	go manager.Timer(AppID, AppSecret, AccessToken)
	// --------------------------------------------------------------- //

	// ----------------创建一个名为"开发者"的标签分组------------------- //
	// 查找是否存在名为"开发者"的标签分组，如果有则返回标签tagid
	id := manager.FindTagToName(AccessToken, TagName)
	if id == 0 { // 表示未找到此name的tag
		// 如果没有则创建标签,并返回标签id
		tagId, err = manager.CreateTag(AccessToken, TagName)
		if err != nil {
			log.Fatal("craete tag error:", err)
		}
		fmt.Println("创建的标签id,tagId=", tagId) // 调试用
	} else { // 存在这个标签，则把返回的标签id存储到全局变量tagId中
		tagId = id
		fmt.Println("存在查找的名字的标签,tagid=", tagId) // 调试用
	}

	// -------------------------------------------------------------- //

	http.HandleFunc("/", procRequest)
	err = http.ListenAndServe(":8088", nil) // 设置的服务器上nginx反代理监听端口为8088，但是server和微信端交互的端口还是80
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
