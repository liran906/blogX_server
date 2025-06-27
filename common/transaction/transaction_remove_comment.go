// Path: ./common/transaction/transaction_remove_comment.go

package transaction

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/comment_service"
	"blogX_server/service/redis_service/redis_article"
	"blogX_server/service/redis_service/redis_comment"
	"fmt"
	"gorm.io/gorm"
)

func RemoveComment(cmt *models.CommentModel) error {
	return global.DBMaster.Transaction(func(tx *gorm.DB) error {
		// 删除子评论
		offsprings, err := comment_service.GetOffsprings(&cmt.ID)
		if err != nil {
			return fmt.Errorf("查询子评论失败, Error: %v", err)
		}
		if len(offsprings) > 0 {
			if err = tx.Delete(&offsprings).Error; err != nil {
				return fmt.Errorf("删除子评论失败, Error: %v", err)
			}
		}

		// 删除本体
		err = tx.Delete(cmt).Error
		if err != nil {
			return fmt.Errorf("删除评论失败, Error: %v", err)
		}

		// 取当前缓存评论数
		// 如果缓存中没有（已经备份到 db），则会返回 0，也没问题
		currentReplyCount := redis_comment.GetCommentReplyCount(cmt.ID) + cmt.ReplyCount

		// 更新祖先评论评论数
		if cmt.ParentID != nil {
			ancestors, err := comment_service.GetAncestors(*cmt.ParentID)
			if err != nil {
				return fmt.Errorf("获取父评论失败, Error: %v", err)
			}
			for _, ans := range ancestors {
				redis_comment.UpdateCommentReplyCount(ans.ID, -currentReplyCount-1)
			}
		}

		// 更新文章评论数
		redis_article.UpdateArticleComment(cmt.ArticleID, -currentReplyCount-1)
		return nil
	})
}
