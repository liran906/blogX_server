// Path: ./api/article_api/article_collection_folder_remove.go

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

func (ArticleApi) ArticleCollectionFolderRemoveView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDListRequest)
	claims := jwts.MustGetClaimsFromRequest(c)

	if len(req.IDList) == 0 {
		res.FailWithMsg("没有指定删除对象", c)
		return
	}

	var list []models.CollectionFolderModel
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
	log.SetTitle("批量删除收藏夹失败")

	var succ []uint
	for _, coll := range list {
		if coll.UserID != claims.UserID {
			if claims.Role != enum.AdminRoleType {
				// 非管理员只能删除自己的记录
				log.SetItem("权限不足", fmt.Sprintf("收藏夹[id: %d][title: %s][belongs to: %d]", coll.ID, coll.Title, coll.UserID))
				continue
			}
			log.ShowClaim(claims) // 管理员删别人
		}
		// 删除事务
		err = transaction.RemoveCollectionFolderTx(&coll)
		if err != nil {
			log.SetItem("失败", fmt.Sprintf("收藏夹[id: %d][title: %s]", coll.ID, coll.Title))
			continue
		}
		log.SetItem("成功", fmt.Sprintf("收藏夹[id: %d][title: %s]", coll.ID, coll.Title))
		succ = append(succ, coll.ID)
	}

	if len(succ) == 0 {
		res.FailWithMsg("批量删除收藏夹失败", c)
		return
	}
	log.SetTitle(fmt.Sprintf("批量删除收藏夹成功"))
	res.SuccessWithMsg(fmt.Sprintf("批量删除收藏夹完成，共计 %d 条，删除 %d 条: %v", len(list), len(succ), succ), c)
}
