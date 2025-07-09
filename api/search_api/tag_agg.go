// Path: ./api/search_api/tag_agg.go

package search_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

// TagAggResponse 定义了接口返回的结构：每个标签及其对应的文章数量
type TagAggResponse struct {
	Tag          string `json:"tag"`
	ArticleCount int    `json:"articleCount"`
}

// TagAggView 是一个标签聚合接口
// 功能：统计当前博客系统中所有已使用的标签（tags），并返回每个标签下的文章数量（可分页）
// 当 Elasticsearch 可用时，使用 terms 聚合实现高性能的标签统计与分页排序
// 当 ES 不可用时，退回使用数据库进行手动聚合（效率较低，仅适用于小数据集）
func (SearchApi) TagAggView(c *gin.Context) {
	var req = c.MustGet("bindReq").(common.PageInfo)
	var list = make([]TagAggResponse, 0)

	// 如果 ES 未连接，使用数据库进行 fallback 查询
	if global.ESClient == nil {
		type ArticleTags struct {
			Tags string `gorm:"column:tags"` // 改为 string
		}

		var results []ArticleTags

		if err := global.DB.Model(&models.ArticleModel{}).
			Select("tags").
			Where("tags <> ''").
			Find(&results).Error; err != nil {
			res.Fail(err, "数据库查询失败", c)
			return
		}

		tagMap := make(map[string]int)
		for _, r := range results {
			tags := parsePGArray(r.Tags)
			for _, tag := range tags {
				tagMap[tag]++
			}
		}

		for tag, count := range tagMap {
			list = append(list, TagAggResponse{
				Tag:          tag,
				ArticleCount: count,
			})
		}

		// 根据文章数量降序排序
		sort.Slice(list, func(i, j int) bool {
			return list[i].ArticleCount > list[j].ArticleCount
		})

		if len(list) > req.Limit {
			list = list[:req.Limit]
		}

		res.SuccessWithList(list, len(list), c)
		return
	}

	// 使用 Elasticsearch 聚合标签（tags 字段）

	// 1. 创建 terms 聚合，按 tags 字段聚合
	agg := elastic.NewTermsAggregation().Field("tags")

	// 2. 添加子聚合：bucket_sort 聚合用于分页（类似 SQL 的 offset/limit）
	// 注意：ES 的 terms 聚合默认返回最多 10 个桶（buckets），如果不加这个分页聚合会被截断
	agg.SubAggregation("page",
		elastic.NewBucketSortAggregation().
			From(req.GetOffset()). // 起始偏移量
			Size(req.Limit))       // 页大小

	// 3. 构建 bool 查询：排除 tags 为空的文章
	query := elastic.NewBoolQuery()
	query.MustNot(elastic.NewTermQuery("tags", ""))

	// 4. 执行搜索请求
	result, err := global.ESClient.
		Search(models.ArticleModel{}.GetIndex()). // 指定索引
		Query(query).                             // 查询条件
		Aggregation("tags", agg).                 // 添加 tags 聚合
		Aggregation("tags1",                      // tags1 是辅助聚合，用于统计标签总数（去重计数）
				elastic.NewCardinalityAggregation().Field("tags")).
		Size(0). // 不返回实际文档，只关注聚合结果
		Do(context.Background())

	if err != nil {
		logrus.Errorf("查询失败 %s", err)
		res.FailWithMsg("查询失败", c)
		return
	}

	// 解析主聚合结果（terms 聚合）
	var t AggType
	var val = result.Aggregations["tags"]
	err = json.Unmarshal(val, &t)
	if err != nil {
		logrus.Errorf("解析json失败 %s %s", err, string(val))
		res.FailWithMsg("查询失败", c)
		return
	}

	// 解析辅助聚合结果（cardinality 聚合）
	var co Agg1Type
	json.Unmarshal(result.Aggregations["tags1"], &co)

	// 构建返回列表
	for _, bucket := range t.Buckets {
		list = append(list, TagAggResponse{
			Tag:          bucket.Key,
			ArticleCount: bucket.DocCount,
		})
	}

	// 返回分页列表和总标签数
	res.SuccessWithList(list, co.Value, c)
	return
}

// AggType 用于反序列化 terms 聚合的结果结构
type AggType struct {
	DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
	SumOtherDocCount        int `json:"sum_other_doc_count"`
	Buckets                 []struct {
		Key      string `json:"key"`       // 标签名
		DocCount int    `json:"doc_count"` // 出现次数（文章数量）
	} `json:"buckets"`
}

// Agg1Type 用于反序列化 cardinality 聚合结果（标签总数）
type Agg1Type struct {
	Value int `json:"value"` // 标签去重后的总数量
}

func parsePGArray(s string) []string {
	s = strings.Trim(s, "{}") // 去掉首尾的 {}
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
