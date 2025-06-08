// Path: ./api/article_api/article_collection_remove.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

func (ArticleApi) ArticleCollectionRemoveView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.RemoveRequest)
	claims := jwts.MustGetClaimsFromGin(c)

	var collections []models.ArticleCollectionModel
	err := global.DB.Preload("ArticleModel").Find(&collections, "id IN ?", req.IDList).Error
	if err != nil {
		res.Fail(err, "数据库读取失败", c)
		return
	}

	if len(collections) == 0 {
		res.FailWithMsg("没有找到可取消收藏的文章", c)
		return
	}

	for _, collection := range collections {
		if collection.UserID != claims.UserID {
			res.FailWithMsg("只能取消自己收藏文章的收藏", c)
			return
		}
		if collection.CollectionFolderID != collections[0].CollectionFolderID {
			res.FailWithMsg("一次只能取消同一个收藏夹内文章的收藏", c)
			return
		}
	}

	err = transaction.RemoveCollectionsTx(collections)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle("批量取消收藏")

	res.SuccessWithMsg("取消收藏成功", c)
}
