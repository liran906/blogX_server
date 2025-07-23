// Path: ./service/ai_service/enter.go

package ai_service

import (
	"blogX_server/global"
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

const baseURL = "https://api.chatanywhere.tech/v1/chat/completions"

//go:embed prompt_chat.prompt
var chatPrompt string

//go:embed prompt_summarize.prompt
var summarizePrompt string

//go:embed prompt_stream.prompt
var streamPrompt string

//go:embed prompt_autogen.prompt
var autoGenPrompt string

type AIChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type requestType int8

const (
	chatAiRequest      requestType = 1
	summarizeAiRequest requestType = 2
	streamAiRequest    requestType = 3
	autogenAiRequest   requestType = 4
)

func baseRequest(msg string, reqType requestType) (res *http.Response, err error) {
	method := "POST"

	var prompt string
	var stream = false
	switch reqType {
	case chatAiRequest:
		prompt = chatPrompt
	case summarizeAiRequest:
		prompt = summarizePrompt
	case autogenAiRequest:
		prompt = autoGenPrompt
	case streamAiRequest:
		prompt = streamPrompt
		stream = true
	}

	var m = AIChatRequest{
		Model: "gpt-4o-mini", // 使用高性价比模型
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
