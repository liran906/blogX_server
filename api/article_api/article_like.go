// Path: ./api/article_api/article_like.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
	"strings"
)

func (ArticleApi) ArticleLikeView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)
	if req.ID == 0 {
		res.FailWithMsg("未指定文章 ID", c)
		return
	}

	var a models.ArticleModel
	err := global.DB.Preload("UserModel").Take(&a, req.ID).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}

	uid := jwts.MustGetClaimsFromGin(c).UserID

	var al models.ArticleLikesModel
	err = global.DB.Take(&al, "article_id = ? and user_id = ?", a.ID, uid).Error
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			err = global.DB.Create(&models.ArticleLikesModel{
				ArticleID: a.ID,
				UserID:    uid,
			}).Error
			if err != nil {
				res.Fail(err, "点赞失败", c)
				return
			}
			res.SuccessWithMsg("点赞成功", c)
			return
			// TODO redis文章点赞数+1
		}
		res.Fail(err, "读取点赞数据失败", c)
	}
	err = global.DB.Delete(&al).Error
	if err != nil {
		res.Fail(err, "取消点赞失败", c)
		return
	}
	res.SuccessWithMsg("取消点赞成功", c)
	// TODO redis文章点赞数-1
}
