// Path: ./service/text_service/enter.go

package text_service

import (
	"fmt"
	"strings"
)

type TextStruct struct {
	ArticleID uint   `json:"article_id"` // 注意这里 json 标签要和 es 一样
	Head      string `json:"head"`
	Body      string `json:"body"`
}

// MDContentTransformation 把一段 md 格式的 article 对象转换为 分段标题+分段内容 格式的 textModel 列表
func MDContentTransformation(aid uint, title, content string) (list []TextStruct) {
	var heads []string
	var bodies []string
	var body string
	var flag bool // 针对代码块内的 #

	heads = append(heads, title)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			flag = !flag
		}
		if !flag && strings.HasPrefix(line, "#") {
			bodies = append(bodies, formatBody(body))
			body = ""
			heads = append(heads, formatHead(line))
			continue
		}
		body += line + "\n"
	}
	if body != "" {
		bodies = append(bodies, formatBody(body))
	}

	// 最后一行是标题，且没有换行，会造成 body 少一个，所以手动补上
	if strings.HasPrefix(lines[len(lines)-1], "#") {
		bodies = append(bodies, "")
	}

	if len(bodies) != len(heads) {
		fmt.Println("文章分段数量不一致：len(bodies) != len(heads)")
		fmt.Printf("%q   %d\n", heads, len(heads))
		fmt.Printf("%q   %d\n", bodies, len(bodies))
		return
	}

	for i := 0; i < len(heads); i++ {
		list = append(list, TextStruct{
			ArticleID: aid,
			Head:      heads[i],
			Body:      bodies[i],
		})
	}
	return
}

func formatHead(head string) string {
	return strings.TrimSpace(strings.Join(strings.Split(head, " ")[1:], " "))
}

func formatBody(body string) string {
	return strings.TrimSpace(body)
}
