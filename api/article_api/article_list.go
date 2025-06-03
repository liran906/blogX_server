// Path: ./api/article_api/article_list.go

package article_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

type ArticleListReq struct {
	common.PageInfo
	UserID          uint               `form:"userID"`
	CategoryID      *uint              `form:"categoryID"`
	Status          enum.ArticleStatus `form:"status"`
	CollectionQuery bool               `form:"viewCollect"` // 查看收藏
}

type ArticleListResp struct {
	models.ArticleModel
}

func (ArticleApi) ArticleListView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleListReq)

	var queryType int8 = 1 // 1 未登录 2 已登录查别人 3已登录查自己 4 管理员

	// 提取身份信息，判断查询种类
	claims, err := jwts.ParseTokenFromRequest(c)
	if err == nil && claims != nil {
		// 不指定就查自己
		if req.UserID == 0 {
			req.UserID = claims.UserID
		}
		// 判断queryTpye
		if claims.Role == enum.AdminRoleType {
			queryType = 4
		} else if req.UserID == claims.UserID {
			queryType = 3
		} else {
			queryType = 2
		}
	}

	// 不登录不指定 id
	if req.UserID == 0 {
		res.FailWithMsg("未指定查询 id", c)
		return
	}
	var u models.UserModel
	err = global.DB.Preload("UserConfigModel").Take(&u, req.UserID).Error
	if err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	switch queryType {
	case 1: // 未登录
		req.Status = enum.ArticleStatusPublish
		req.CollectionQuery = false
		if req.PageInfo.Page > 1 || req.PageInfo.Limit > 10 {
			res.FailWithMsg("登录后查看更多", c)
			return
		}
	case 2: // 已登录查别人
		if !u.UserConfigModel.DisplayCollections {
			req.CollectionQuery = false
		}
		req.Status = enum.ArticleStatusPublish
		// 情况 3 和 4 都是最高权限，不做限制
	}

	if !req.CollectionQuery {
		// 发布文章查询
		_list, count, err := common.ListQuery(
			models.ArticleModel{
				UserID:     req.UserID,
				CategoryID: req.CategoryID,
				Status:     req.Status,
			},
			common.Options{
				PageInfo: req.PageInfo,
				Likes:    []string{"title"},
				Debug:    true,
			})
		if err != nil {
			res.FailWithMsg("查询失败", c)
			return
		}
		res.SuccessWithList(_list, count, c)
	} else {
		// 收藏文章查询
	}
}
