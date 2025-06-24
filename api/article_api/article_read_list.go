// Path: ./api/article_api/article_read_list.go

package article_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
	"time"
)

type ArticleReadListReq struct {
	common.PageInfo
	UserId uint `form:"userID"`
}

type ArticleReadListResp struct {
	HistoryID uint      `json:"historyID"` // 浏览记录的 id
	CreatedAt time.Time `json:"createdAt"`
	UserId    uint      `json:"userID"`
	ArticleId uint      `json:"articleID"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatarURL"`
	Title     string    `json:"title"`
	CoverURL  string    `json:"coverURL"`
	Abstract  string    `json:"abstract"`
}

func (ArticleApi) ArticleReadListView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleReadListReq)
	claims := jwts.MustGetClaimsFromRequest(c)

	// 没有传入就查自己
	if req.UserId == 0 {
		req.UserId = claims.UserID
	}

	// 非管理员只能查自己
	if claims.UserID != req.UserId && claims.Role != enum.AdminRoleType {
		res.FailWithMsg("只能查看自己的浏览历史", c)
		return
	}

	req.PageInfo.Normalize()

	// 解析时间戳并查询
	query, err := common.TimeQuery(req.StartTime, req.EndTime)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	// 查询
	_list, count, err := common.ListQuery(
		models.UserArticleHistoryModel{UserID: req.UserId},
		common.Options{
			PageInfo: req.PageInfo,
			Where:    query,
			Preloads: []string{"UserModel", "ArticleModel"},
		})
	if err != nil {
		res.Fail(err, "查询失败", c)
		return
	}
	if len(_list) == 0 {
		res.FailWithMsg("没有找到匹配记录", c)
		return
	}

	// 构造响应
	list := make([]ArticleReadListResp, 0, len(_list))
	for _, uh := range _list {
		list = append(list, ArticleReadListResp{
			HistoryID: uh.ID,
			UserId:    uh.UserID,
			ArticleId: uh.ArticleID,
			CreatedAt: uh.CreatedAt,
			Nickname:  uh.UserModel.Nickname,
			AvatarURL: uh.UserModel.AvatarURL,
			Title:     uh.ArticleModel.Title,
			CoverURL:  uh.ArticleModel.CoverURL,
			Abstract:  uh.ArticleModel.Abstract,
		})
	}
	res.SuccessWithList(list, count, c)
}
