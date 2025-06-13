// Path: ./api/ai_api/article_ai.go

package ai_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/ai_service"
	"blogX_server/service/log_service"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"strings"
)

type ArticleAiReq struct {
	Content string `form:"content" binding:"required"`
}

func (AiApi) ArticleAiView(c *gin.Context) {
	if !global.Config.Ai.Enable {
		res.FailWithMsg("站点未开启 AI 功能", c)
		return
	}

	req := c.MustGet("bindReq").(ArticleAiReq)

	var match string
	if global.ESClient == nil {
		// 未开启 es，用 mysql 搜索
		list, _, err := common.ListQuery(&models.ArticleModel{
			Status: enum.ArticleStatusPublish,
		}, common.Options{
			Likes: []string{"title", "abstract"},
			PageInfo: common.PageInfo{
				Page:  1,
				Limit: 5, // 限制 5 条
			},
		})
		if err != nil {
			res.Fail(err, "搜索失败", c)
			return
		}
		byteData, err := json.Marshal(list)
		if err != nil {
			res.Fail(err, "搜索失败", c)
			return
		}
		match = string(byteData)
	} else {
		// 采用 es 搜索
		// 创建一个布尔查询对象，用于组合多个查询条件
		query := elastic.NewBoolQuery()

		// 1. Must（必须匹配，类似 SQL 中的 AND）
		// status = 3 表示已发布的文章
		// NewTermQuery 用于精确匹配，不会对查询词进行分词
		query.Must(elastic.NewTermQuery("status", 3))

		query.Should(
			elastic.NewMatchQuery("title", req.Content),
			elastic.NewMatchQuery("abstract", req.Content),
			elastic.NewMatchQuery("content", req.Content),
		)

		result, err := global.ESClient.
			Search(models.ArticleModel{}.GetIndex()). // 搜索的是哪一个 index
			Query(query).                             // 什么类型的查询以及具体查询条件
			From(0).                                  // 从哪一条开始显示
			Size(5).                                  // 往后显示多少条
			Do(context.Background())                  // 执行
		if err != nil {
			source, _ := query.Source()
			byteData, _ := json.Marshal(source)
			logrus.Errorf("查询失败 %s \n %s", err, string(byteData))
			res.Fail(err, "查询失败", c)
			return
		}

		var list []string
		for _, hit := range result.Hits.Hits {
			list = append(list, string(hit.Source))
		}
		match = "json data: [" + strings.Join(list, ",") + "]"
	}

	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle("AI 查询")
	log.SetItemTrace("搜索结果: ", match)

	msg := "json data: " + match + "\n" + "user query: " + "\"" + req.Content + "\""
	msgChan, err := ai_service.ChatStream(msg)
	if err != nil {
		res.Fail(err, "AI响应失败", c)
		return
	}

	// 流式回复用 sse 响应。go 实现就两行。
	// 请求数据不大的前提下，用 get 请求，这样前端会比较好写
	for s := range msgChan {
		res.SSESuccess(s, c)
	}
}
