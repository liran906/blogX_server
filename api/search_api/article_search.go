// Path: ./api/search_api/article_search.go

package search_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/redis_service/redis_article"
	"blogX_server/utils/jwts"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"strings"
)

type ArticleSearchReq struct {
	common.PageInfo
	Type int8   `form:"type" binding:"oneof=0 1 2 3 4 5 6"` // 0-猜你喜欢 1-最新发布 2-最多回复 3-最多点赞 4-最多收藏 5-最多阅读量 6-最新更新
	Tag  string `form:"tag"`
}

type ArticleBaseInfo struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
}

type ArticleListResp struct {
	models.ArticleModel
	UserNickname  string  `json:"userNickname,omitempty"`
	UserAvatarURL string  `json:"userAvatarURL,omitempty"`
	CategoryName  *string `json:"categoryName,omitempty"`
}

func (SearchApi) ArticleSearchView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleSearchReq)
	req.PageInfo.Normalize()

	claims, err := jwts.ParseTokenFromRequest(c)
	if err != nil || claims == nil {
		// 未登录状态，只能看一页
		if req.PageInfo.Page > 1 || req.PageInfo.Limit > 10 {
			res.FailWithMsg("登录查看更多", c)
			return
		}
	}

	// 搜索顺序判断
	var sortMap = map[int8]string{
		0: "_score",        // 猜你喜欢
		1: "created_at",    // 最新发布
		2: "comment_count", // 最多回复
		3: "like_count",    // 最多点赞
		4: "collect_count", // 最多收藏
		5: "read_count",    // 最多阅读量
		6: "updated_at",    // 最新更新
	}
	sortKey := sortMap[req.Type]

	// 读取缓存中的数据
	readMap := redis_article.GetAllReadCounts()
	likeMap := redis_article.GetAllLikeCounts()
	collectMap := redis_article.GetAllCollectCounts()
	commentMap := redis_article.GetAllCommentCounts()

	// 没有开启 ES，也能实现服务降级（用 mysql）的搜索
	if global.ESClient == nil {
		var defaultOrder string
		if req.Type == 0 {
			// 这里的 type 0 就是置顶在前，其他按创建日期排序
			// 找出置顶文章 id 列表
			var pinnedArticleIDList []uint
			err = global.DB.Model(&models.UserPinnedArticleModel{UserID: 0}).Order("`rank` ASC").Pluck("article_id", &pinnedArticleIDList).Error
			if err != nil {
				res.Fail(err, "读取置顶文章失败", c)
				return
			}
			// 置顶在前
			for _, aid := range pinnedArticleIDList {
				defaultOrder += fmt.Sprintf("id = %d DESC, ", aid)
			}
			defaultOrder += "created_at DESC"
		} else {
			defaultOrder = sortMap[req.Type] + " DESC"
		}

		where := global.DB.Where("")
		if req.Tag != "" {
			where = where.Where("tags LIKE ?", "%"+req.Tag+"%")
		}

		// 解析时间戳并查询
		where, err = common.TimeQueryWithBase(where, req.StartTime, req.EndTime)
		if err != nil {
			res.FailWithMsg(err.Error(), c)
			return
		}

		_list, count, _ := common.ListQuery(models.ArticleModel{
			Status: enum.ArticleStatusPublish,
		}, common.Options{
			PageInfo:     req.PageInfo,
			Preloads:     []string{"UserModel", "CategoryModel"},
			Likes:        []string{"title", "abstract"}, // 这里不考虑正文了
			Where:        where,
			DefaultOrder: defaultOrder,
			Debug:        false,
		})
		var list []ArticleListResp
		for _, a := range _list {
			a.ReadCount += readMap[a.ID]
			a.LikeCount += likeMap[a.ID]
			a.CollectCount += collectMap[a.ID]
			a.CommentCount += commentMap[a.ID]
			item := ArticleListResp{
				ArticleModel:  a,
				UserNickname:  a.UserModel.Nickname,
				UserAvatarURL: a.UserModel.AvatarURL,
			}
			if a.CategoryModel != nil {
				item.CategoryName = &a.CategoryModel.Name
			}
			list = append(list, item)
		}
		res.SuccessWithList(list, count, c)
		return
	}

	// 以下是正常开启了 ES 的服务：
	// 创建一个布尔查询对象，用于组合多个查询条件
	query := elastic.NewBoolQuery()

	// 1. Must（必须匹配，类似 SQL 中的 AND）
	// status = 3 表示已发布的文章
	// NewTermQuery 用于精确匹配，不会对查询词进行分词
	query.Must(elastic.NewTermQuery("status", 3))

	// 2. 如果指定了标签，添加标签过滤条件
	// 标签也使用 Must 确保强制匹配（AND）
	if req.Tag != "" {
		query.Must(
			// NewTermQuery 用于精确匹配标签，因为标签通常是固定词
			elastic.NewTermQuery("tags", req.Tag),
		)
	}

	highlight := elastic.NewHighlight()

	// 3. 关键词搜索（Should 条件，提高相关性评分）
	if req.Key != "" {
		// Should 条件类似 SQL 中的 OR
		// 匹配越多的条件，文档的相关性评分越高
		query.Should(
			// NewMatchQuery 会对查询词进行分词，更适合全文搜索
			// 以下三个字段都会参与搜索，任一匹配即可
			elastic.NewMatchQuery("title", req.Key),    // 标题匹配
			elastic.NewMatchQuery("abstract", req.Key), // 摘要匹配
			elastic.NewMatchQuery("content", req.Key),  // 内容匹配
		)
		// 注：可以通过 Boost() 方法调整各字段的权重
		// 例如：elastic.NewMatchQuery("title", req.Key).Boost(3) 让标题匹配的权重更高

		// 设置高亮显示
		highlight.Field("title")
		highlight.Field("abstract")
	} else {
		// 没有搜索 key，才会按照个人 tag 搜索
		// 查询type 为“猜你喜欢”，并且登录了
		if req.Type == 0 && claims != nil {
			// 找用户感兴趣的标签
			var uc models.UserConfigModel
			err = global.DB.Take(&uc, "user_id = ?", claims.UserID).Error
			if err != nil {
				res.Fail(err, "读取用户配置失败", c)
				return
			}

			if len(uc.Tags) > 0 {
				var shouldQueries []elastic.Query
				for _, tag := range uc.Tags {
					shouldQueries = append(shouldQueries, elastic.NewMatchQuery("title", tag))
					//query.Should(elastic.NewMatchQuery("tittle", tag))
					//query.Should(elastic.NewMatchQuery("abstract", tag))
				}
				query.Should(shouldQueries...)
			}
		}
	}

	result, err := global.ESClient.
		Search(models.ArticleModel{}.GetIndex()). // 搜索的是哪一个 index
		Query(query). // 什么类型的查询以及具体查询条件
		Highlight(highlight). // 高亮关键词
		From(req.GetOffset()). // 从哪一条开始显示
		Size(req.GetLimit()). // 往后显示多少条
		Sort(sortKey, false). // 排序
		Do(context.Background()) // 执行
	if err != nil {
		res.Fail(err, "查询失败", c)
		return
	}

	count := int(result.Hits.TotalHits.Value)      // 获取搜索结果的总条数
	searchResult := make(map[uint]ArticleBaseInfo) // 创建一个 map 用于存储搜索结果，key 为文章 ID，value 为文章基本信息
	var esSortedIDList []uint                      // 创建一个由 ES 算法排序的文章 idList

	for _, hit := range result.Hits.Hits { // 遍历每一个搜索命中的文档
		var abi ArticleBaseInfo                // 创建文章基本信息对象
		err = json.Unmarshal(hit.Source, &abi) // 将 ES 文档源数据（_source）解析为 ArticleBaseInfo 结构体
		if err != nil {
			logrus.Errorf("json 解析失败: %v", err) // 如果解析失败，记录错误
			continue                                // 继续处理下一条
		}

		// 如果存在标题的高亮结果，使用高亮后的标题替换原标题
		if len(hit.Highlight["title"]) > 0 {
			abi.Title = hit.Highlight["title"][0] // 高亮结果是一个数组，取第一个元素
		}

		// 如果存在摘要的高亮结果，使用高亮后的摘要替换原摘要
		if len(hit.Highlight["abstract"]) > 0 {
			abi.Abstract = hit.Highlight["abstract"][0] // 高亮结果是一个数组，取第一个元素
		}

		searchResult[abi.ID] = abi                      // 将处理后的文章信息存入结果 map
		esSortedIDList = append(esSortedIDList, abi.ID) // (按顺序)保存搜索出来的文章 id
	}

	// TODO ==｜关于最终的输出顺序｜==
	// TODO
	// TODO type 为 0 的时候，直接按照 es 搜索出来的次序（当然，还要考虑管理员置顶）
	// TODO 其他的 type 则用 mysql 排序（因为 es 的排序无法实时同步 redis 中的数据）
	// TODO 其实 es 最大的作用就是模糊搜索时候的快速匹配，所以非模糊的时候（type1-6）用处没那么大
	// TODO 还有个作用就是高亮显示，只要有 key 就会对应高亮
	var defaultOrder string
	// 如果是进入网站主页（type 是 0，没有 key，也没有 tag）那么管理员置顶的优先展示
	if req.Type == 0 {
		if req.Key == "" && req.Tag == "" {
			// 找出置顶文章 id 列表
			var pinnedArticleIDList []uint
			err = global.DB.Model(&models.UserPinnedArticleModel{}).Where("user_id = ?", 0).
				Order("`rank` ASC").Pluck("article_id", &pinnedArticleIDList).Error
			if err != nil {
				res.Fail(err, "读取置顶文章失败", c)
				return
			}
			// 将置顶文章 id 放到最前面
			esSortedIDList = append(pinnedArticleIDList, esSortedIDList...)
		}

	} else {
		// type 为 1-6，顺序先读取 redis 的数据在排序
	}

	// 根据 esSortedIDList 中的顺序，写出 SQL 的排序语句
	for _, aid := range esSortedIDList {
		defaultOrder += fmt.Sprintf("id = %d DESC, ", aid)
	}
	defaultOrder = strings.TrimSuffix(defaultOrder, ", ") // 修个尾巴

	// 查询 db
	where := global.DB.Where("id IN ?", esSortedIDList)

	// 解析时间戳并查询
	where, err = common.TimeQueryWithBase(where, req.StartTime, req.EndTime)
	if err != nil {
		res.Fail(err, "时间解析失败", c)
		return
	}

	_list, _, _ := common.ListQuery(models.ArticleModel{}, common.Options{
		Preloads:     []string{"UserModel", "CategoryModel"},
		Where:        where,
		DefaultOrder: defaultOrder,
		Debug:        false,
	})

	var list []ArticleListResp
	for _, a := range _list {
		a.ReadCount += readMap[a.ID]
		a.LikeCount += likeMap[a.ID]
		a.CollectCount += collectMap[a.ID]
		a.CommentCount += commentMap[a.ID]
		item := ArticleListResp{
			ArticleModel:  a,
			UserNickname:  a.UserModel.Nickname,
			UserAvatarURL: a.UserModel.AvatarURL,
		}
		if a.CategoryModel != nil {
			item.CategoryName = &a.CategoryModel.Name
		}
		item.Title = searchResult[a.ID].Title
		item.Abstract = searchResult[a.ID].Abstract
		list = append(list, item)
	}
	res.SuccessWithList(list, count, c)
}
