// Path: ./api/article_api/article_tag_options.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/ctype"
	"blogX_server/models/enum"
	"blogX_server/utils"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

func (ArticleApi) ArticleTagOptionsView(c *gin.Context) {
	claims := jwts.MustGetClaimsFromGin(c)

	// 找到所有发布过的文章
	var alist []models.ArticleModel
	err := global.DB.Find(&alist, "user_id = ? AND status = ?", claims.UserID, enum.ArticleStatusPublish).Error
	if err != nil {
		res.Fail(err, "查询数据库失败", c)
		return
	}

	// 找到过去所有发布文章用过的标签
	var tagList ctype.List
	for _, article := range alist {
		tagList = append(tagList, article.Tags...)
	}

	// 标签去重
	tagList = utils.Unique(tagList)

	// 生成返回结果
	var list []models.OptionsRequest[string]
	for _, tag := range tagList {
		list = append(list, models.OptionsRequest[string]{
			Label: tag,
			Value: tag,
		})
	}
	res.SuccessWithData(list, c)
}
