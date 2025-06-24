// Path: ./api/article_api/article_collection_folder_list.go

package article_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/utils/jwts"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleCollectionFolderListReq struct {
	common.PageInfo
	UserID uint `form:"userID"`
}

func (ArticleApi) ArticleCollectionFolderListView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCollectionFolderListReq)
	claims := jwts.MustGetClaimsFromRequest(c)

	// 不指定就查自己
	if req.UserID == 0 {
		req.UserID = claims.UserID
	}

	// 不查自己要校验
	if req.UserID != claims.UserID {
		// 查别人需要验证 id 是否存在
		var u models.UserModel
		err := global.DB.Preload("UserConfigModel").Take(&u, req.UserID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				res.Fail(err, "用户不存在", c)
				return
			}
			res.Fail(err, "查询数据库失败", c)
			return
		}
		// 校验是否公开收藏夹
		if claims.Role != enum.AdminRoleType && !u.UserConfigModel.DisplayCollections {
			res.FailWithMsg("对方未公开收藏夹", c)
			return
		}
	}

	req.PageInfo.Normalize()

	// 解析时间戳并查询
	query, err := common.TimeQuery(req.StartTime, req.EndTime)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	list, count, err := common.ListQuery(models.CollectionFolderModel{
		UserID: req.UserID,
	}, common.Options{
		PageInfo: req.PageInfo,
		Likes:    []string{"title"},
		Where:    query,
		Debug:    false,
	})
	if err != nil {
		res.Fail(err, "查询失败", c)
		return
	}
	res.SuccessWithList(list, count, c)
}
