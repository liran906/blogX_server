// Path: ./api/article_api/article_category_list.go

package article_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleCategoryListReq struct {
	common.PageInfo
	UserID    uint   `form:"userID"`
	StartTime string `form:"startTime"` // format "2006-01-02 15:04:05"
	EndTime   string `form:"endTime"`
}

type ArticleCategoryListResp struct {
	models.CategoryModel
	ArticleCount int `json:"articleCount"`
}

func (ArticleApi) ArticleCategoryListView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCategoryListReq)

	// 这里和教程不一样，我觉得首先必须登录，其次并不需要判断是否是管理员
	// 按教程：管理员多返回一个 nickname 和 avatar 有什么意义呢？uid 都是有的。

	var u models.UserModel
	err := global.DB.Take(&u, req.UserID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Fail(err, "用户不存在", c)
			return
		}
		res.Fail(err, "查询数据库失败", c)
		return
	}

	req.PageInfo.Normalize()

	// 解析时间戳并查询
	query, err := common.TimeQuery(req.StartTime, req.EndTime)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	_list, count, err := common.ListQuery(models.CategoryModel{
		UserID: req.UserID,
	}, common.Options{
		PageInfo: req.PageInfo,
		Preloads: []string{"ArticleList"},
		Where:    query,
		Debug:    false,
	})
	if err != nil {
		res.Fail(err, "查询失败", c)
		return
	}

	list := make([]ArticleCategoryListResp, 0, len(_list))
	for _, cat := range _list {
		list = append(list, ArticleCategoryListResp{
			CategoryModel: cat,
			ArticleCount:  len(cat.ArticleList),
		})
	}
	res.SuccessWithList(list, count, c)
}
