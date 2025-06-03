// Path: ./utils/xss/enter.go

package xss

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
)

func Filter(content string) string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(content)))
	if err != nil {
		fmt.Println("文本解析错误")
		return ""
	}
	doc.Find("script").Remove()
	doc.Find("iframe").Remove()
	doc.Find("img").Remove()

	return doc.Text()
}
