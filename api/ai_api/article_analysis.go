// Path: ./api/ai_api/article_analysis.go

package ai_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models/enum"
	"blogX_server/service/ai_service"
	"blogX_server/service/log_service"
	"encoding/json"
	"github.com/gin-gonic/gin"
)

type ArticleAnalysisReq struct {
	Content string `json:"content" binding:"required"`
}

type ArticleAnalysisResp struct {
	Title    string   `json:"title"`
	Abstract string   `json:"abstract"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

func (AiApi) ArticleAnalysisView(c *gin.Context) {
	if !global.Config.Ai.Enable {
		res.FailWithMsg("站点未开启 AI 功能", c)
		return
	}

	req := c.MustGet("bindReq").(ArticleAnalysisReq)
	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()

	resp, err := ai_service.Chat(req.Content)
	if err != nil {
		log.SetLevel(enum.LogWarnLevel)
		log.SetTitle("AI分析失败")
		log.SetItemWarn(err.Error(), resp)
		res.Fail(err, "AI 分析失败", c)
		return
	}

	var data ArticleAnalysisResp
	err = json.Unmarshal([]byte(resp), &data)
	if err != nil {
		log.SetLevel(enum.LogWarnLevel)
		log.SetTitle("AI响应解析失败")
		log.SetItemWarn(err.Error(), resp)
		res.Fail(err, "AI 分析失败", c)
		return
	}
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle("AI分析成功")
	res.SuccessWithData(data, c)
}
