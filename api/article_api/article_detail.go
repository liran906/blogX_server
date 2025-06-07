// Path: ./api/article_api/article_detail.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/redis_service/redis_article"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

type ArticleDetailResp struct {
	models.ArticleModel
	UserNickname  string `json:"userNickname"`
	UserAvatarURL string `json:"userAvatarURL"`
}

func (ArticleApi) ArticleDetailView(c *gin.Context) {
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
	_ = redis_article.UpdateCachedFieldsForArticle(&a) // 读取缓存数据

	// 提取身份信息，判断查询种类
	claims, err := jwts.ParseTokenFromRequest(c)
	// 未登录无法看未发布的文章
	if err != nil && a.Status != enum.ArticleStatusPublish {
		res.FailWithMsg("文章不存在", c)
		return
	}
	if err == nil && claims != nil {
		// 非管理员无法看别人未发布的文章
		if claims.Role != enum.AdminRoleType && a.UserID != claims.UserID && a.Status != enum.ArticleStatusPublish {
			res.FailWithMsg("文章不存在", c)
			return
		}
	}

	resp := ArticleDetailResp{
		ArticleModel:  a,
		UserNickname:  a.UserModel.Nickname,
		UserAvatarURL: a.UserModel.AvatarURL,
	}
	res.SuccessWithData(resp, c)
}
