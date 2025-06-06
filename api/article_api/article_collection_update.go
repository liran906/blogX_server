// Path: ./api/article_api/article_collection_update.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
	"strings"
)

type ArticleCollectionUpdateReq struct {
	CollectionID uint   `json:"collectionID" binding:"required"`
	Title        string `json:"title"`
	Abstract     string `json:"abstract"`
	CoverURL     string `json:"coverURL"`
}

func (ArticleApi) ArticleCollectionUpdateView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCollectionUpdateReq)
	claims := jwts.MustGetClaimsFromGin(c)

	var cf models.CollectionFolderModel
	err := global.DB.Take(&cf, req.CollectionID).Error
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
		"abstract": req.Abstract,
		"coverURL": req.CoverURL,
	}
	if req.Title != "" {
		updateMap["title"] = req.Title
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
