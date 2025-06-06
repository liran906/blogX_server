// Path: ./api/article_api/article_category_create.go

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

type ArticleCategoryCreateReq struct {
	Name string `json:"name" binding:"required,max=32"`
}

func (ArticleApi) ArticleCategoryCreateView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCategoryCreateReq)
	claims := jwts.MustGetClaimsFromGin(c)

	var cm models.CategoryModel
	err := global.DB.Where("user_id = ? AND name = ?", claims.UserID, req.Name).Take(&cm).Error
	if err == nil {
		res.FailWithMsg("分类已存在", c)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		res.Fail(err, "查询数据库失败", c)
		return
	}

	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle("创建文章分类")

	err = global.DB.Create(&models.CategoryModel{
		UserID: claims.UserID, Name: req.Name,
	}).Error
	if err != nil {
		res.Fail(err, "分类创建失败", c)
		return
	}
	res.SuccessWithMsg("分类创建成功", c)
}
