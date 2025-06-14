// Path: ./api/article_api/article_read_remove.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
)

func (ArticleApi) ArticleReadRemoveView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDListRequest)
	claims := jwts.MustGetClaimsFromRequest(c)

	if len(req.IDList) == 0 {
		res.FailWithMsg("没有指定删除对象", c)
		return
	}

	var list []models.UserArticleHistoryModel
	err := global.DB.Where("id IN ?", req.IDList).Find(&list).Error
	if err != nil {
		res.Fail(err, "查询数据库失败", c)
		return
	}

	// 校验查出来的记录数量是否和请求一致
	if len(list) != len(req.IDList) {
		res.FailWithMsg("部分记录不存在或无权限访问", c)
		return
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("删除浏览历史")

	// 非管理员只能删除自己的记录
	for _, uh := range list {
		if uh.UserID != claims.UserID {
			if claims.Role != enum.AdminRoleType {
				res.FailWithMsg("只能删除自己的浏览记录", c)
				return
			}
			log.ShowClaim(claims) // 管理员删别人
			break
		}
	}

	// 删除
	err = global.DB.Delete(&list).Error
	if err != nil {
		res.Fail(err, "删除失败", c)
		return
	}

	// redis 记录不删除 不然能刷阅读量了
	//for _, uh := range list {
	//	redis_article.RemoveUserArticleHistoryCacheToday(uh.ArticleID, uh.UserID)
	//}

	res.SuccessWithMsg(fmt.Sprintf("成功删除 %d 条", len(list)), c)
}
