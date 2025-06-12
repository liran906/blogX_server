// Path: ./service/ai_service/enter.go

package ai_service

import (
	"blogX_server/global"
	_ "embed"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

const baseURL = "https://api.chatanywhere.tech/v1/chat/completions"

//go:embed prompt_chat.prompt
var chatPrompt string

//go:embed prompt_stream.prompt
var streamPrompt string

type AIChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func baseRequest(msg string, stream bool) (res *http.Response, err error) {
	method := "POST"

	prompt := chatPrompt
	if stream {
		prompt = streamPrompt
	}

	var m = AIChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "system",
				Content: prompt,
			},
			{
				Role:    "user",
				Content: msg,
			},
		},
		Stream: stream,
	}
	bd, err := json.Marshal(m)
	if err != nil {
		logrus.Errorf("json 解析失败 %s", err)
		return
	}
	payload := strings.NewReader(string(bd))
	//payload1 := bytes.NewBuffer(bd)

	req, err := http.NewRequest(method, baseURL, payload)
	if err != nil {
		logrus.Errorf("请求解析失败 %s", err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+global.Config.Ai.SecretKey)
	req.Header.Add("Content-Type", "application/json")

	res, err = http.DefaultClient.Do(req)
	return
}
