package test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
)

func main01() {
	var t xml.Token

	var err error
	// input := `<Person><FirstName>Xu</FirstName><LastName>Xinhang</LastName></Person>`
	// inputReader := strings.NewReader(input)
	// decoder := xml.NewDecoder(inputReader)
	content, _ := ioutil.ReadFile("studyGolang.xml")
	decoder := xml.NewDecoder(bytes.NewBuffer(content))
	for t, err = decoder.Token(); err == nil; t, err = decoder.Token() {
		switch token := t.(type) {
		// 处理元素开始（标签）
		case xml.StartElement:
			name := token.Name.Local
			fmt.Printf("token name:%s\n", name)
			for _, attr := range token.Attr {
				attrName := attr.Name.Local
				attrValue := attr.Value
				fmt.Printf("An attribute is: %s %s \n", attrName, attrValue)
			}
			// 处理元素结束（标签）
		case xml.EndElement:
			fmt.Printf("Token of '%s' end\n", token.Name.Local)
			// 处理字符数据(这里就是元素的文本)
		case xml.CharData:
			content := string([]byte(token))
			fmt.Printf("this is the content: %v\n", content)
		default:
			// ...
		}
	}
}
func main02() {
	content, err := ioutil.ReadFile("studyGolang.xml")
	if err != nil {
		log.Fatal(err)
	}
	var result Result
	err = xml.Unmarshal(content, &result)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(result)
}

type Result struct {
	Person []Person
}
type Person struct {
	Name      string `xml:"name,attr"`
	Age       int    `xml:"age,attr"`
	Career    string
	Interests Interests
}
type Interests struct {
	Interest []string
}
