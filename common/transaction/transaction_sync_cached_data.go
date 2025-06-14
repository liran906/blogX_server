// Path: ./common/transaction/transaction_sync_cached_data.go

package transaction

import (
	"blogX_server/global"
	"blogX_server/models"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strconv"
)

func SyncCommentTx(commentList []models.CommentModel, maps map[string]map[uint]int) error {
	replyMap := maps["reply"]
	likeMap := maps["like"]
	return global.DBMaster.Transaction(func(tx *gorm.DB) error {
		for _, cmt := range commentList {
			// 遍历每一篇文章提取出更新数据
			updateMap := make(map[string]any, 2)
			if d, ok := replyMap[cmt.ID]; ok && d != 0 {
				updateMap["reply_count"] = gorm.Expr("reply_count + ?", d)
			}
			if d, ok := likeMap[cmt.ID]; ok && d != 0 {
				updateMap["like_count"] = gorm.Expr("like_count + ?", d)
			}

			// 如果有全为 0 的情况，上面 activeComment 是无法筛选出来的，所以这里再筛一次
			if len(updateMap) == 0 {
				continue
			}

			// 写入数据库
			err := tx.Model(&cmt).Updates(updateMap).Error
			if err != nil {
				return fmt.Errorf("update comment[%d] error: %v", cmt.ID, err)
			}
			logrus.Infof("update comment[%d]", cmt.ID)
		}
		return nil
	})
}

func SyncArticleTx(articleList []models.ArticleModel, maps map[string]map[uint]int) error {
	readMap := maps["read"]
	likeMap := maps["like"]
	collectMap := maps["collect"]
	commentMap := maps["comment"]
	return global.DBMaster.Transaction(func(tx *gorm.DB) error {
		// 这里是遍历数据库查出来的文章，所以也不怕 redis 中有数据，而 db 被删了的情况
		for _, article := range articleList {
			// 遍历每一篇文章提取出更新数据
			updateMap := make(map[string]any, 4)
			if d, ok := readMap[article.ID]; ok && d != 0 {
				updateMap["read_count"] = gorm.Expr("read_count + ?", d)
			}
			if d, ok := likeMap[article.ID]; ok && d != 0 {
				updateMap["like_count"] = gorm.Expr("like_count + ?", d)
			}
			if d, ok := collectMap[article.ID]; ok && d != 0 {
				updateMap["collect_count"] = gorm.Expr("collect_count + ?", d)
			}
			if d, ok := commentMap[article.ID]; ok && d != 0 {
				updateMap["comment_count"] = gorm.Expr("comment_count + ?", d)
			}

			// 如果有全为 0 的情况，上面 activeArticle 是无法筛选出来的，所以这里再筛一次
			if len(updateMap) == 0 {
				continue
			}

			// 写入数据库
			err := tx.Model(&article).Updates(updateMap).Error
			if err != nil {
				return fmt.Errorf("update article[%d] error: %v", article.ID, err)
			}
			logrus.Infof("update article[%d]", article.ID)
		}
		return nil
	})
}

func SyncUserTx(mps map[string]string) error {
	return global.DBMaster.Transaction(func(tx *gorm.DB) error {
		for k, v := range mps {
			uid, _ := strconv.Atoi(k)
			num, _ := strconv.Atoi(v)
			err := global.DB.Where("user_id=?", uid).Model(&models.UserConfigModel{}).Update("homepage_visit_count", gorm.Expr("homepage_visit_count + ?", num)).Error
			if err != nil {
				return fmt.Errorf("update user[%s] error: %v", k, err)
			}
		}
		return nil
	})
}
