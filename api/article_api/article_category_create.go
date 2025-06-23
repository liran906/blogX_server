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
	"strings"
)

type ArticleCategoryCreateReq struct {
	ID    uint   `json:"id"`
	Title string `json:"title" binding:"required,max=32"`
}

// ArticleCategoryCreateView 再次，为了迎合前端，把创建和更新都放入这个 api 了
// 有 id 就是更新，没有 id （id==0）就是创建
func (ArticleApi) ArticleCategoryCreateView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCategoryCreateReq)
	claims := jwts.MustGetClaimsFromRequest(c)

	// 新建逻辑
	if req.ID == 0 {
		var cm models.CategoryModel
		err := global.DB.Where("user_id = ? AND name = ?", claims.UserID, req.Title).Take(&cm).Error
		if err == nil {
			res.FailWithMsg("分类已存在", c)
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			res.Fail(err, "查询数据库失败", c)
			return
		}

		log := log_service.GetActionLog(c)
		log.ShowRequest()
		log.ShowResponse()
		log.SetLevel(enum.LogTraceLevel)
		log.SetTitle("创建文章分类")

		err = global.DB.Create(&models.CategoryModel{
			UserID: claims.UserID, Name: req.Title,
		}).Error
		if err != nil {
			res.Fail(err, "分类创建失败", c)
			return
		}
		res.SuccessWithMsg("分类创建成功", c)
		return
	}

	// 更新逻辑
	var cm models.CategoryModel
	err := global.DB.Take(&cm, req.ID).Error
	if err != nil {
		res.Fail(err, "查询数据库失败", c)
		return
	}

	if req.Title == cm.Name {
		res.SuccessWithMsg("分类更新成功", c)
		return
	}

	if claims.Role != enum.AdminRoleType && claims.UserID != cm.UserID {
		res.FailWithMsg("只能修改自己的分类", c)
		return
	}

	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle("更新文章分类")

	err = global.DB.Model(&cm).Update("name", req.Title).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			res.Fail(err, "分类已存在", c)
			return
		}
		res.Fail(err, "分类更新失败", c)
		return
	}
	res.SuccessWithMsg("分类更新成功", c)
}
