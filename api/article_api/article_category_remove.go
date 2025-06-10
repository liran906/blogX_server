// Path: ./api/article_api/article_category_remove.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
)

func (ArticleApi) ArticleCategoryRemoveView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDListRequest)
	claims := jwts.MustGetClaimsFromGin(c)

	if len(req.IDList) == 0 {
		res.FailWithMsg("没有指定删除对象", c)
		return
	}

	var list []models.CategoryModel
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
	log.SetTitle("批量删除文章分类失败")

	var succ []uint
	for _, cat := range list {
		if cat.UserID != claims.UserID {
			if claims.Role != enum.AdminRoleType {
				// 非管理员只能删除自己的记录
				log.SetItem("权限不足", fmt.Sprintf("文章分类[id: %d][name: %s][belongs to: %d]", cat.ID, cat.Name, cat.UserID))
				continue
			}
			log.ShowClaim(claims) // 管理员删别人
		}
		err = transaction.RemoveCategory(&cat)
		if err != nil {
			log.SetItem("失败", fmt.Sprintf("文章分类[id: %d][name: %s]", cat.ID, cat.Name))
			continue
		}
		log.SetItem("成功", fmt.Sprintf("文章分类[id: %d][name: %s]", cat.ID, cat.Name))
		succ = append(succ, cat.ID)
	}

	if len(succ) == 0 {
		res.FailWithMsg("批量删除分类失败", c)
		return
	}
	log.SetTitle(fmt.Sprintf("批量删除文章分类成功"))
	res.SuccessWithMsg(fmt.Sprintf("批量删除分类完成，共计 %d 条，删除 %d 条: %v", len(list), len(succ), succ), c)
}
