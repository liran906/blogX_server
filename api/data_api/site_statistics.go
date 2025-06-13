// Path: ./api/data_api/site_statistics.go

package data_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/redis_service/redis_article"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

type SiteStatisticsReq struct {
	Duration int    `json:"duration" binding:"required"` // 几就是几天
	EndDate  string `json:"endDate"`
}

type SiteStatisticsResp struct {
	StartTime          time.Time `json:"startTime"`
	EndTime            time.Time `json:"endTime"`
	TotalUserCount     int64     `json:"totalUserCount"`
	TotalArticleCount  int64     `json:"totalArticleCount"`
	FlowCount          int64     `json:"flowCount"`
	LoginCount         int64     `json:"loginCount"`
	RegisterCount      int64     `json:"registerCount"`
	NewArticleCount    int64     `json:"newArticleCount"`
	ActiveArticleCount int64     `json:"activeArticleCount"`
	CommentCount       int64     `json:"commentCount"`
	LikeCount          int64     `json:"likeCount"`
	CollectCount       int64     `json:"collectCount"`
}

func (DataApi) SiteStatisticsView(c *gin.Context) {
	req := c.MustGet("bindReq").(SiteStatisticsReq)
	if req.Duration < 1 || req.Duration > 365 {
		res.FailWithMsg("只支持查询 1-365 天范围内数据", c)
		return
	}

	// 处理结束时间
	now := time.Now()
	var endDate time.Time
	if req.EndDate == "" {
		// 获取明天的日期 0 点
		endDate = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	} else {
		var err error
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			// 如果解析失败，使用明天日期 0 点
			now := time.Now()
			endDate = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		} else {
			// 如果成功解析，也要加一天
			endDate = endDate.AddDate(0, 0, 1)
		}
	}
	// 计算开始时间
	startDate := time.Date(endDate.Year(), endDate.Month(), endDate.Day()-req.Duration, 0, 0, 0, 0, endDate.Location())
	// 创建时间查询条件
	between := "created_at BETWEEN ? AND ?"

	var resp SiteStatisticsResp
	resp.StartTime = startDate
	resp.EndTime = endDate
	global.DB.Model(&models.UserModel{}).Count(&resp.TotalUserCount)                                                                        // 所有用户
	global.DB.Model(&models.ArticleModel{}).Count(&resp.TotalArticleCount)                                                                  // 所有文章
	global.DB.Where(between, startDate, endDate).Where("log_type = ?", enum.LoginLogType).Model(&models.LogModel{}).Count(&resp.LoginCount) // 新增登录（不考虑 jwt 的过期了）
	global.DB.Where(between, startDate, endDate).Model(&models.UserModel{}).Count(&resp.RegisterCount)                                      // 新增注册用户
	global.DB.Where(between, startDate, endDate).Model(&models.ArticleModel{}).Count(&resp.NewArticleCount)                                 // 新增文章
	global.DB.Where(between, startDate, endDate).Model(&models.ArticleLikesModel{}).Count(&resp.LikeCount)                                  // 新增点赞
	global.DB.Where(between, startDate, endDate).Model(&models.ArticleCollectionModel{}).Count(&resp.CollectCount)                          // 新增收藏
	global.DB.Where(between, startDate, endDate).Model(&models.CommentModel{}).Count(&resp.CommentCount)                                    // 新增评论

	// 下面都是在统计活跃文章
	activeArticleIds := make(map[uint]struct{})
	// 分别查询并存入 map
	var articleIds []uint

	// 查询点赞文章
	global.DB.Where(between, startDate, endDate).Model(&models.ArticleLikesModel{}).
		Distinct().Pluck("article_id", &articleIds)
	for _, id := range articleIds {
		activeArticleIds[id] = struct{}{}
	}
	// 查询收藏文章
	articleIds = nil // 清空切片
	global.DB.Where(between, startDate, endDate).Model(&models.ArticleCollectionModel{}).
		Distinct().Pluck("article_id", &articleIds)
	for _, id := range articleIds {
		activeArticleIds[id] = struct{}{}
	}
	// 查询评论文章
	articleIds = nil // 清空切片
	global.DB.Where(between, startDate, endDate).Model(&models.CommentModel{}).
		Distinct().Pluck("article_id", &articleIds)
	for _, id := range articleIds {
		activeArticleIds[id] = struct{}{}
	}

	// 如果包含今天，要加上 redis 数据
	if endDate == time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()) {
		mps := global.Redis.HGetAll(string(redis_article.ArticleLikeCount)).Val()
		for k, v := range mps {
			val, _ := strconv.Atoi(v)
			resp.LikeCount += int64(val)
			key, _ := strconv.Atoi(k)
			activeArticleIds[uint(key)] = struct{}{}
		}
		mps = global.Redis.HGetAll(string(redis_article.ArticleCollectCount)).Val()
		for k, v := range mps {
			val, _ := strconv.Atoi(v)
			resp.CollectCount += int64(val)
			key, _ := strconv.Atoi(k)
			activeArticleIds[uint(key)] = struct{}{}
		}
		mps = global.Redis.HGetAll(string(redis_article.ArticleCommentCount)).Val()
		for k, v := range mps {
			val, _ := strconv.Atoi(v)
			resp.CommentCount += int64(val)
			key, _ := strconv.Atoi(k)
			activeArticleIds[uint(key)] = struct{}{}
		}
	}
	resp.ActiveArticleCount = int64(len(activeArticleIds))
	res.Success(resp, "成功", c)
}
