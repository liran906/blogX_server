// Path: ./api/article_api/article_collect.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/service/redis_service/redis_article"
	"blogX_server/utils/jwts"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strings"
)

type ArticleCollectReq struct {
	ArticleID    uint `json:"articleID" binding:"required"`
	CollectionID uint `json:"collectionID"`
}

// ArticleCollectView 收藏某篇文章
func (ArticleApi) ArticleCollectView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCollectReq)
	uid := jwts.MustGetClaimsFromGin(c).UserID
	var cf models.CollectionFolderModel

	// 只允许收藏已发布状态的文章
	var a models.ArticleModel
	err := global.DB.Take(&a, "id = ? AND status = ?", req.ArticleID, 3).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}

	// 没有填写收藏夹 id
	if req.CollectionID == 0 {
		// 查找用户默认收藏夹
		err := global.DB.Take(&cf, "user_id = ? and is_default = true", uid).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 没有默认收藏夹，创建
				cf.UserID = uid
				cf.Title = "默认收藏夹"
				cf.IsDefault = true
				err := global.DB.Create(&cf).Error
				if err != nil {
					res.Fail(err, "默认收藏夹创建失败", c)
					return
				}
			} else {
				res.Fail(err, "查询数据库失败", c)
				return
			}
		}
		req.CollectionID = cf.ID
	} else {
		// 查询收藏夹是否存在
		err := global.DB.Take(&cf, "id = ? AND user_id = ?", req.CollectionID, uid).Error
		if err != nil {
			res.Fail(err, "收藏夹不存在", c)
			return
		}
	}

	ac := models.ArticleCollectionModel{
		ArticleID:          req.ArticleID,
		UserID:             uid,
		CollectionFolderID: req.CollectionID,
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle(fmt.Sprintf("增加收藏 %d", req.ArticleID))

	// userID articleID collectionFolderID 是一组联合 CK，如果有重复写入会自己报错
	// 所以不在这里显示判断是否有重复
	err = global.DB.Create(&ac).Error
	if err != nil {
		// 判断是否是已经有记录了
		if strings.Contains(err.Error(), "Duplicate entry") {
			// 已有记录，取消收藏
			err := global.DB.Delete(&ac, "user_id = ? AND article_id = ? AND collection_folder_id = ?", ac.UserID, ac.ArticleID, ac.CollectionFolderID).Error
			if err != nil {
				res.Fail(err, "取消收藏失败", c)
				return
			}
			// 更新数量
			global.DB.Model(&cf).Update("article_count", gorm.Expr("article_count - 1"))
			redis_article.SubArticleCollect(req.ArticleID)
			log.SetTitle(fmt.Sprintf("取消收藏 %d", req.ArticleID))
			res.SuccessWithMsg("取消收藏成功", c)
			return
		} else {
			res.Fail(err, "查询数据库失败", c)
			return
		}
	}
	// 更新数量
	global.DB.Model(&cf).Update("article_count", gorm.Expr("article_count + 1"))
	redis_article.AddArticleCollect(req.ArticleID)
	res.SuccessWithMsg("收藏成功", c)
}
