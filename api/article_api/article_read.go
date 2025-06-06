// Path: ./api/article_api/article_read.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/redis_service/redis_article"
	"blogX_server/utils/jwts"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type ArticleCountReadReq struct {
	ArticleID uint `json:"articleID" binding:"required"`
	Interval  uint `json:"interval"` // TODO 秒单位
}

// ArticleCountReadView 增加浏览记录
func (ArticleApi) ArticleCountReadView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCountReadReq)
	var a models.ArticleModel
	err := global.DB.Take(&a, "id = ? AND status = ?", req.ArticleID, 3).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}

	claims, err := jwts.ParseTokenFromRequest(c)
	if err != nil || claims == nil {
		// TODO 未登录逻辑
		res.SuccessWithMsg("未登录，成功", c)
		return
	}

	if redis_article.HasUserArticleHistoryCacheToday(a.ID, claims.UserID) {
		res.SuccessWithMsg("今日已有足迹", c)
		return
	}

	// 其实下面关于 mysql 的判断都可以省略，记录全部存入缓存，每天固定时间同步即可。
	// 不过这里就按照教程吧，还是写入 mysql

	// 查这个文章今天有没有在足迹里面
	var his models.UserArticleHistoryModel
	err = global.DB.Take(&his,
		"user_id = ? AND article_id = ? AND created_at < ? AND created_at >= ?",
		claims.UserID,
		req.ArticleID,
		time.Now().AddDate(0, 0, 1).Format("2006-01-02")+" 00:00:00",
		time.Now().Format("2006-01-02")+" 00:00:00",
	).Error
	if err == nil {
		// 有足迹
		res.SuccessWithMsg("今日已有足迹!", c)
		return
	}
	// 没有足迹
	if errors.Is(err, gorm.ErrRecordNotFound) {
		his.ArticleID = req.ArticleID
		his.UserID = claims.UserID
		err := global.DB.Create(&his).Error
		if err != nil {
			res.Fail(err, "创建失败", c)
			return
		}
		// 写入逻辑
		redis_article.SetUserArticleHistoryCacheToday(a.ID, claims.UserID)
		redis_article.AddArticleRead(req.ArticleID)
		res.SuccessWithMsg("成功", c)
		return
	} else {
		res.Fail(err, "查询数据库失败", c)
		return
	}
}
