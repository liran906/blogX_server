// Path: ./api/article_api/article_category_options.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

// ArticleCategoryOptionsView 是用于**获取当前登录用户的文章分类选项列表（options）**的
// 常用于前端页面上的 下拉选择框（select），比如在用户写文章时选择“属于哪个分类”
func (ArticleApi) ArticleCategoryOptionsView(c *gin.Context) {
	claims := jwts.MustGetClaimsFromRequest(c)

	//var cm []models.CategoryModel
	//err := global.DB.Find(&cm, "user_id=?", claims.UserID).Error
	//if err != nil {
	//	res.Fail(err, "数据库查询失败", c)
	//	return
	//}
	//var resp []models.OptionsRequest[uint]
	//for _, v := range cm {
	//	resp = append(resp, models.OptionsRequest[uint]{
	//		Label: v.Title,
	//		Value: v.ID,
	//	})
	//}

	// 上面的可以用下面代替 等效。
	var resp []models.OptionsRequest[uint]
	err := global.DB.Model(&models.CategoryModel{}).Where("user_id = ?", claims.UserID).
		Select("id AS value", "name AS label").Scan(&resp).Error
	if err != nil {
		res.Fail(err, "数据库查询失败", c)
		return
	}
	res.SuccessWithData(resp, c)
}
