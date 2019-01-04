package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// access_token 返回的数据结构类型
type AccToken struct {
	Token     string `json:"access_token"`
	LimitTime int    `json:"expires_in"`
}

// 批量为用户打标签
type PlayTag struct {
	OpenidList []string `json:"openid_list"`
	Tagid      int      `json:"tagid"`
}

// unmarshal 创建标签的数据结构
type UnmarshalTag struct {
	Tag *TagParam `json:"tag"`
}
type TagParam struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// marshal 创建标签的数据结构
type MarshalTag struct {
	Tag *TagName `json:"tag"`
}
type TagName struct {
	Name string `json:"name"`
}

// 获取公众号已创建的标签的数据结构
type FindTagName struct {
	Tags []TagMsg `json:"tags"`
}
type TagMsg struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// 1.获取到公众号的access_token 注：access_token有效期为2小时，access_token是调用微信端api的唯一识别码。
func GetAccessToken(appId, appSecret string) (string, error) {
	Url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appId, appSecret)
	resp, err := http.Get(Url)
	if err != nil {
		log.Println("get access_token error:", err)
		return "", err
	}
	byteGet, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println("ioutil reader error:", err)
		return "", err
	}
	token := &AccToken{}
	json.Unmarshal(byteGet, token)

	return token.Token, nil
}

// 2.创建一个标签,name为创建标签的名字。并返回创建标签的id, 每个标签只能创建一次。
func CreateTag(accessToken, name string) (int, error) {
	Url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/tags/create?access_token=%s", accessToken)
	// 请求数据处理
	marshalData := &MarshalTag{
		Tag: &TagName{
			Name: name,
		},
	}
	data, err := json.Marshal(marshalData)
	if err != nil {
		log.Println("marshal error:", err)
		return 0, err
	}
	reader := bytes.NewReader(data)
	resp, err := http.Post(Url, "application/json;charset=UTF-8", reader)
	if err != nil {
		log.Println("post error:", err)
		return 0, err
	}
	res, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	// 反序列化
	tag := &UnmarshalTag{Tag: &TagParam{}}
	json.Unmarshal(res, tag)

	fmt.Printf("创建标签返回的 id=%d name=%s\n", tag.Tag.Id, tag.Tag.Name) // 调试用
	return tag.Tag.Id, nil
}

// 3.为申请打币的用户添加进 id = "开发者"标签 的分组，accessToken为access_token，openIds为用户的open id集合，id 为标签id，由微信分配
func AddTagForUser(accessToken string, openIds []string, id int) error {
	Url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/tags/members/batchtagging?access_token=%s", accessToken)
	// 请求的数据
	playtag := &PlayTag{
		OpenidList: openIds,
		Tagid:      id,
	}
	byteData, err := json.Marshal(playtag)
	if err != nil {
		log.Println("json marshal play tag error:", err)
		return err
	}
	reader := bytes.NewReader(byteData)
	resp, err := http.Post(Url, "application/json;charset=UTF-8", reader)
	if err != nil {
		log.Println("play tag post error:", err)
		return err
	}

	// 调试用,打印出响应的数据，打印出 {"errcode":0,"errmsg":"ok"} 则表示操作成功
	res, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println("批量为用户打标签post响应数据：", string(res))

	return nil
}

// 4.定时器函数，定时每2小时更新access_token,AccessToken为更新到的全局变量
func Timer(appId, appSecret, AccessToken string) {
	tick := time.NewTicker(7140 * time.Second) // 提前1分钟进行更新
	for {
		select {
		case <-tick.C:
			var err error
			AccessToken, err = GetAccessToken(appId, appSecret)
			if err != nil {
				log.Println("get access_token error:", err)
				return
			}
			fmt.Println("定时更新后的access_token:", AccessToken) // 调试用
		}
	}
}

// 5.查找公众号已创建的标签中是否存在给定name的标签,如果存在则返回标签对应的tagid,如果不存在则返回0
func FindTagToName(AccessToken, name string) int {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/tags/get?access_token=%s", AccessToken)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("获取公众号已创建的标签失败:", err)
	}
	res, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	// 反序列化得到findTagName的结构体
	findTagName := &FindTagName{Tags: []TagMsg{}}
	json.Unmarshal(res, findTagName)

	// 找出是否有名字为name的tag
	for i := 0; i < len(findTagName.Tags); i++ {
		if findTagName.Tags[i].Name == name {
			return findTagName.Tags[i].Id
		}
	}
	return 0
}
