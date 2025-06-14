// Path: ./api/article_api/article_category_update.go

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

type ArticleCategoryUpdateReq struct {
	CategoryId uint   `json:"categoryID" binding:"required"`
	Name       string `json:"name" binding:"required,max=32"`
}

func (ArticleApi) ArticleCategoryUpdateView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCategoryUpdateReq)
	claims := jwts.MustGetClaimsFromRequest(c)

	var cm models.CategoryModel
	err := global.DB.Take(&cm, req.CategoryId).Error
	if err != nil {
		res.Fail(err, "查询数据库失败", c)
		return
	}

	if req.Name == cm.Name {
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

	err = global.DB.Model(&cm).Update("name", req.Name).Error
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
