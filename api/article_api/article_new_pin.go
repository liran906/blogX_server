// Path: ./api/article_api/article_new_pin.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/redis_service/redis_cache"
	"blogX_server/utils/jwts"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ArticleNewAdminPinView 做前端的时候，置顶是一个开关，所以新写了这个方法，一次置顶（或者取消置顶）一篇文章
// 由于没法传入 rank 了，所以通过这个方法置顶的文章的 rank 全部是 1，排序就根据时间倒序（越新置顶的越靠上）
// 管理员置顶和普通用户置顶，通过 uid 区分，管理员置顶 uid == 0；普通用户置顶，uid 就是用户的 id
func (ArticleApi) ArticleNewAdminPinView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)

	var a models.ArticleModel
	err := global.DB.Find(&a, req.ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Fail(err, "文章不存在", c)
			return
		}
		res.Fail(err, "数据库查询失败", c)
		return
	}

	if a.Status != enum.ArticleStatusPublish {
		res.FailWithMsg("只能置顶已发布的文章", c)
		return
	}

	err = transaction.NewPinArticleTx(0, &a)
	if err != nil {
		res.Fail(err, "修改置顶失败", c)
		return
	}
	redis_cache.CacheCloseCertain(fmt.Sprintf("%s%d", redis_cache.CacheArticleDetailPrefix, a.ID))
	res.SuccessWithMsg("修改置顶成功", c)
}

// ArticleNewUserPinView 做前端的时候，置顶是一个开关，所以新写了这个方法，一次置顶（或者取消置顶）一篇文章
// 由于没法传入 rank 了，所以通过这个方法置顶的文章的 rank 全部是 1，排序就根据时间倒序（越新置顶的越靠上）
// 管理员置顶和普通用户置顶，通过 uid 区分，管理员置顶 uid == 0；普通用户置顶，uid 就是用户的 id
func (ArticleApi) ArticleNewUserPinView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)
	claims := jwts.MustGetClaimsFromRequest(c)

	var a models.ArticleModel
	err := global.DB.Find(&a, req.ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Fail(err, "文章不存在", c)
			return
		}
		res.Fail(err, "数据库查询失败", c)
		return
	}

	if a.Status != enum.ArticleStatusPublish {
		res.FailWithMsg("只能置顶已发布的文章", c)
		return
	}

	err = transaction.NewPinArticleTx(claims.UserID, &a)
	if err != nil {
		res.Fail(err, "修改置顶失败", c)
		return
	}
	redis_cache.CacheCloseCertain(fmt.Sprintf("%s%d", redis_cache.CacheArticleDetailPrefix, a.ID))
	res.SuccessWithMsg("修改置顶成功", c)
}
