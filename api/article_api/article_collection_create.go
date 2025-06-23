// Path: ./api/article_api/article_collection_create.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strings"
)

type ArticleCollectionCreateReq struct {
	ID       uint   `json:"id"`
	Title    string `json:"title" binding:"required"`
	Abstract string `json:"abstract"`
	CoverURL string `json:"coverURL"`
}

// ArticleCollectionFolderCreateView 再次，为了迎合前端，把创建和更新都放入这个 api 了
// 有 id 就是更新，没有 id （id==0）就是创建
func (ArticleApi) ArticleCollectionFolderCreateView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCollectionCreateReq)
	claims := jwts.MustGetClaimsFromRequest(c)
	var cf models.CollectionFolderModel

	// 创建逻辑
	if req.ID == 0 {
		err := global.DB.Where("user_id = ? AND title = ?", claims.UserID, req.Title).Take(&cf).Error
		if err == nil {
			res.FailWithMsg("收藏夹已存在", c)
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			res.Fail(err, "查询数据库失败", c)
			return
		}

		// 日志
		log := log_service.GetActionLog(c)
		log.ShowRequest()
		log.ShowResponse()
		log.SetLevel(enum.LogTraceLevel)
		log.SetTitle("创建收藏夹")

		// 创建
		err = global.DB.Create(&models.CollectionFolderModel{
			UserID:   claims.UserID,
			Title:    req.Title,
			Abstract: req.Abstract,
			CoverURL: req.CoverURL,
		}).Error
		if err != nil {
			res.Fail(err, "收藏夹创建失败", c)
			return
		}
		res.SuccessWithMsg("收藏夹创建成功", c)
		return
	}

	// 更新逻辑
	err := global.DB.Take(&cf, req.ID).Error
	if err != nil {
		res.Fail(err, "查询数据库失败", c)
		return
	}

	if claims.Role != enum.AdminRoleType && claims.UserID != cf.UserID {
		res.FailWithMsg("只能修改自己的收藏夹", c)
		return
	}

	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle("更新收藏夹")

	updateMap := map[string]any{
		"title":     req.Title,
		"abstract":  req.Abstract,
		"cover_url": req.CoverURL,
	}
	err = global.DB.Model(&cf).Updates(updateMap).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			res.Fail(err, "收藏夹已存在", c)
			return
		}
		res.Fail(err, "收藏夹更新失败", c)
		return
	}
	res.SuccessWithMsg("收藏夹更新成功", c)
}
