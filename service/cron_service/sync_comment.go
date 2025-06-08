// Path: ./service/cron_service/sync_comment.go

package cron_service

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/redis_service/redis_comment"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

func SyncComment() {
	// 记录时间
	start := time.Now()

	// 从 redis 中读取数据
	replyMap := redis_comment.GetAllReplyCounts()
	likeMap := redis_comment.GetAllLikeCounts()

	// redis 归零（同步期间如果有增量继续记录，明天再统计）
	redis_comment.Clear()

	// 创建一个字典记录有改变的 cid
	var activeComments = make(map[uint]struct{})
	maps := []map[uint]int{replyMap, likeMap}
	for _, m := range maps {
		for cid := range m { // 只要 key
			activeComments[cid] = struct{}{}
		}
	}

	if len(activeComments) == 0 {
		logrus.Info("no active comments to sync")
		return
	}

	// 从 DB 中取出对应的文章
	var commentList []models.CommentModel
	err := global.DB.Where("id IN ?", mapKeys(activeComments)).Find(&commentList).Error
	if err != nil {
		logrus.Errorf("get article list error: %v", err)
		return
	}

	// 遍历comment，修改数据
	count := 0
	for _, cmt := range commentList {
		// 遍历每一篇文章提取出更新数据
		updateMap := make(map[string]any, 2)
		if d, ok := replyMap[cmt.ID]; ok && d != 0 {
			updateMap["read_count"] = gorm.Expr("read_count + ?", d)
		}
		if d, ok := likeMap[cmt.ID]; ok && d != 0 {
			updateMap["like_count"] = gorm.Expr("like_count + ?", d)
		}

		// 如果有全为 0 的情况，上面 activeComment 是无法筛选出来的，所以这里再筛一次
		if len(updateMap) == 0 {
			continue
		}

		// 写入数据库
		err = global.DB.Model(&cmt).Updates(updateMap).Error
		if err != nil {
			logrus.Errorf("update comment[%d] error: %v", cmt.ID, err)
			continue
		}
		logrus.Infof("update comment[%d]", cmt.ID)
		count++
	}
	logrus.Infof("update comment complete, total %d comments, %d success, %s time elapsed", len(commentList), count, time.Since(start))
}
