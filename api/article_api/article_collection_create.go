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
)

type ArticleCollectionCreateReq struct {
	Title    string `json:"title" binding:"required"`
	Abstract string `json:"abstract"`
	CoverURL string `json:"coverURL"`
}

func (ArticleApi) ArticleCollectionFolderCreateView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCollectionCreateReq)
	claims := jwts.MustGetClaimsFromGin(c)

	var cf models.CollectionFolderModel
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
}
