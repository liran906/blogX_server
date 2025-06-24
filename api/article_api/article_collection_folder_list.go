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
	UserID    uint `form:"userID"`
	ArticleID uint `form:"articleID"` // 传入 articleID 就是问这篇文章，有没有被收藏到返回的收藏夹列表中的某几个收藏夹内
}

type ArticleCollectionFolderListRes struct {
	models.CollectionFolderModel
	//UserNickname  string `json:"userNickname,omitempty"`
	//UserAvatarURL string `json:"userAvatarURL,omitempty"`
	ArticleUse bool `json:"articleUse,omitempty"`
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

	_list, count, err := common.ListQuery(models.CollectionFolderModel{
		UserID: req.UserID,
	}, common.Options{
		PageInfo: req.PageInfo,
		Likes:    []string{"title"},
		Preloads: []string{"UserModel"},
		Where:    query,
		Debug:    false,
	})
	if err != nil {
		res.Fail(err, "查询失败", c)
		return
	}

	// 如果传入了 articleID，查这个 id 是否被收藏过
	var articleCollectList []models.ArticleCollectionModel
	var collectMap = map[uint]struct{}{} // key是收藏夹的 id，有 struct 就代表这个收藏夹收藏了 article id 对应的文章
	if req.ArticleID != 0 && req.UserID == claims.UserID {
		global.DB.Model(models.ArticleCollectionModel{}).Find(&articleCollectList, "article_id = ? AND user_id = ?", req.ArticleID, req.UserID)
		if len(articleCollectList) > 0 {
			for _, collect := range articleCollectList {
				collectMap[collect.CollectionFolderID] = struct{}{}
			}
		}
	}

	list := make([]ArticleCollectionFolderListRes, 0, len(_list))
	for _, cf := range _list {
		resp := ArticleCollectionFolderListRes{
			CollectionFolderModel: cf,
			//UserNickname:          cf.UserModel.Nickname,
			//UserAvatarURL:         cf.UserModel.AvatarURL,
			ArticleUse: false,
		}
		if _, exists := collectMap[cf.ID]; exists {
			resp.ArticleUse = true
		}
		list = append(list, resp)
	}
	res.SuccessWithList(list, count, c)
}
