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
	UserNickname  string  `json:"userNickname"`
	UserAvatarURL string  `json:"userAvatarURL"`
	CategoryName  *string `json:"categoryName"`
	IsLiked       bool    `json:"isLiked"`
	IsCollected   bool    `json:"isCollected"`
}

func (ArticleApi) ArticleDetailView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)
	if req.ID == 0 {
		res.FailWithMsg("未指定文章 ID", c)
		return
	}

	var a models.ArticleModel
	var resp ArticleDetailResp
	err := global.DB.Preload("UserModel").Preload("CategoryModel").Take(&a, req.ID).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}

	// 提取身份信息，判断查询种类
	claims, err := jwts.ParseTokenFromRequest(c)
	// 未登录无法看未发布的文章
	if (err != nil || claims == nil) && a.Status != enum.ArticleStatusPublish {
		res.FailWithMsg("文章不存在", c)
		return
	}
	if err == nil && claims != nil {
		// 非管理员无法看别人未发布的文章
		if claims.Role != enum.AdminRoleType && a.UserID != claims.UserID && a.Status != enum.ArticleStatusPublish {
			res.FailWithMsg("文章不存在", c)
			return
		}
		// 查询文章是否被自己收藏及点赞
		err := global.DB.Take(&models.ArticleCollectionModel{}, "article_id = ? AND user_id = ?", a.ID, claims.UserID).Error
		if err == nil {
			resp.IsCollected = true
		}
		err = global.DB.Take(&models.ArticleLikesModel{}, "article_id = ? AND user_id = ?", a.ID, claims.UserID).Error
		if err == nil {
			resp.IsLiked = true
		}
	}

	// 从 redis 更新点赞收藏等数据
	redis_article.UpdateCachedFieldsForArticle(&a)

	// 更新响应体的其他字段
	resp.ArticleModel = a
	resp.UserNickname = a.UserModel.Nickname
	resp.UserAvatarURL = a.UserModel.AvatarURL

	if a.CategoryModel != nil {
		resp.CategoryName = &a.CategoryModel.Name
	}
	res.SuccessWithData(resp, c)
}
