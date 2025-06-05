// Path: ./common/transaction/transaction_remove_article.go

package transaction

import (
	"blogX_server/global"
	"blogX_server/models"
	"fmt"
	"gorm.io/gorm"
)

func RemoveArticleAndRelated(a *models.ArticleModel) (logs map[string]any, err error) {
	var (
		likes    []models.ArticleLikesModel
		collects []models.ArticleCollectionModel
		pins     []models.UserPinnedArticleModel
		history  []models.UserArticleHistoryModel
		comments []models.CommentModel
	)

	err = global.DBMaster.Transaction(func(tx *gorm.DB) (err error) {
		// 点赞
		if err := tx.Where("article_id = ?", a.ID).Find(&likes).Delete(&models.ArticleLikesModel{}).Error; err != nil {
			return err
		}
		// 收藏
		if err := tx.Where("article_id = ?", a.ID).Find(&collects).Delete(&models.ArticleCollectionModel{}).Error; err != nil {
			return err
		}
		// 置顶
		if err := tx.Where("article_id = ?", a.ID).Find(&pins).Delete(&models.UserPinnedArticleModel{}).Error; err != nil {
			return err
		}
		// 阅读
		if err := tx.Where("article_id = ?", a.ID).Find(&history).Delete(&models.UserArticleHistoryModel{}).Error; err != nil {
			return err
		}
		// 评论
		if err := tx.Where("article_id = ?", a.ID).Find(&comments).Delete(&models.CommentModel{}).Error; err != nil {
			return err
		}
		// 文章本体
		if err := tx.Delete(a).Error; err != nil {
			return err
		}

		logs = map[string]any{
			fmt.Sprintf("删除文章 %d", a.ID):              a,
			fmt.Sprintf("删除关联点赞 %d 条", len(likes)):    likes,
			fmt.Sprintf("删除关联收藏 %d 条", len(collects)): collects,
			fmt.Sprintf("删除关联置顶 %d 条", len(pins)):     pins,
			fmt.Sprintf("删除关联阅读 %d 条", len(history)):  history,
			fmt.Sprintf("删除关联评论 %d 条", len(comments)): comments,
		}
		return nil
	})
	return
}
