// Path: ./api/search_api/text_search.go

package search_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/text_service"
	"blogX_server/utils/jwts"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

type TextSearchReq struct {
	common.PageInfo
	StartTime string `form:"startTime"` // format "2006-01-02 15:04:05"
	EndTime   string `form:"endTime"`
}

// buyongfengzhuangleba?

type TextSearchResp struct {
	ArticleID uint   `json:"articleID"`
	Head      string `json:"head"`
	Body      string `json:"body"`
}

func (SearchApi) TextSearchView(c *gin.Context) {
	req := c.MustGet("bindReq").(TextSearchReq)
	req.PageInfo.Normalize()

	claims, err := jwts.ParseTokenFromRequest(c)
	if err != nil || claims == nil {
		// 未登录状态，只能看一页
		if req.PageInfo.Page > 1 || req.PageInfo.Limit > 10 {
			res.FailWithMsg("登录查看更多", c)
			return
		}
	}

	// 没有开启 ES，也能实现服务降级（用 mysql）的搜索
	if global.ESClient == nil {
		// 解析时间戳并查询
		query, err := common.TimeQuery(req.StartTime, req.EndTime)
		if err != nil {
			res.Fail(err, "时间解析失败", c)
			return
		}

		_list, count, _ := common.ListQuery(models.TextModel{},
			common.Options{
				PageInfo: req.PageInfo,
				Likes:    []string{"head", "body"},
				Where:    query,
				Debug:    false,
			})
		var list []TextSearchResp
		for _, t := range _list {
			item := TextSearchResp{
				ArticleID: t.ArticleID,
				Head:      t.Head,
				Body:      t.Body,
			}
			list = append(list, item)
		}
		res.SuccessWithList(list, count, c)
		return
	}

	// 以下是正常开启了 ES 的服务：
	// 创建一个布尔查询对象，用于组合多个查询条件
	query := elastic.NewBoolQuery()

	// 关键词搜索（Should 条件，提高相关性评分）
	if req.Key != "" {
		query = query.Should(elastic.NewMultiMatchQuery(req.Key, "head", "body"))
	}

	// 设置高亮显示
	highlight := elastic.NewHighlight()
	highlight.Field("head")
	highlight.Field("body")

	result, err := global.ESClient.
		Search(models.TextModel{}.GetIndex()). // 搜索的是哪一个 index
		Query(query).                          // 什么类型的查询以及具体查询条件
		Highlight(highlight).                  // 高亮关键词
		From(req.PageInfo.GetOffset()).        // 从哪一条开始显示
		Size(req.PageInfo.Limit).              // 往后显示多少条
		Do(context.Background())               // 执行
	if err != nil {
		source, _ := query.Source()
		byteData, _ := json.Marshal(source)
		msg := fmt.Sprintf("查询失败 %s \n %s", err, string(byteData))
		logrus.Errorf(msg)
		res.Fail(err, "查询失败", c)
		return
	}

	count := int(result.Hits.TotalHits.Value) // 获取搜索结果的总条数
	var list []TextSearchResp                 // 实例响应体列表

	for _, hit := range result.Hits.Hits {
		var item text_service.TextStruct        // 注意这里用 text_service.TextStruct，因为 articleID 有个小写的解析才行
		err = json.Unmarshal(hit.Source, &item) // 将 ES 文档源数据（_source）解析为 ArticleBaseInfo 结构体
		if err != nil {
			logrus.Errorf("json 解析失败: %v", err) // 如果解析失败，记录错误
			continue                            // 继续处理下一条
		}

		// 如果存在高亮结果，使用高亮后的标题替换原标题
		if len(hit.Highlight["head"]) > 0 {
			item.Head = hit.Highlight["head"][0] // 高亮结果是一个数组，取第一个元素
		}

		// 如果存在高亮结果，使用高亮后的摘要替换原摘要
		if len(hit.Highlight["body"]) > 0 {
			item.Body = hit.Highlight["body"][0] // 高亮结果是一个数组，取第一个元素
		}

		list = append(list, TextSearchResp{
			ArticleID: item.ArticleID,
			Head:      item.Head,
			Body:      item.Body,
		})
	}

	res.SuccessWithList(list, count, c)
}
