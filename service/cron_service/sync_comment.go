// Path: ./service/cron_service/sync_comment.go

package cron_service

import (
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/service/redis_service/redis_comment"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

func SyncComment() {
	// 记录时间
	start := time.Now()

	log := log_service.NewRuntimeLog("同步评论数据", log_service.RuntimeDeltaDay)
	log.SetItem("开始时间", start.Format("2006-01-02 15:04:05"))

	// 从 redis 中读取数据
	replyMap := redis_comment.GetAllReplyCounts()
	likeMap := redis_comment.GetAllLikeCounts()

	// redis 归零（同步期间如果有增量继续记录，明天再统计）
	redis_comment.Clear()

	// 创建一个字典记录有改变的 cid
	var activeComments = make(map[uint]struct{})
	maps := map[string]map[uint]int{"reply": replyMap, "like": likeMap}
	for _, m := range maps {
		for cid := range m { // 只要 key
			activeComments[cid] = struct{}{}
		}
	}

	if len(activeComments) == 0 {
		logrus.Info("no active comments to sync")
		log.SetTitle("无新数据")
		log.Save()
		return
	}

	log.SetTitle("同步失败")

	// 从 DB 中取出本次有修改的评论
	var commentList []models.CommentModel
	err := global.DB.Where("id IN ?", mapKeys(activeComments)).Find(&commentList).Error
	if err != nil {
		logrus.Errorf("get comment list error: %v", err)
		log.SetItemError("查询失败", fmt.Sprintf("get comment list error: %v", err))
		log.SetLevel(enum.LogErrorLevel)
		log.Save()
		return
	}

	// 事务中遍历comment，修改数据
	err = transaction.SyncCommentTx(commentList, maps)
	if err != nil {
		logrus.Errorf("sync comment error: %v", err)
		log.SetItemWarn("事务失败", fmt.Sprintf("sync comment error: %v", err))
		log.SetLevel(enum.LogWarnLevel)
		if err = RollbackCommentRedis(replyMap, likeMap); err != nil {
			logrus.Errorf("rollback to Redis error: %v", err)
			log.SetItemError("回滚失败", fmt.Sprintf("rollback to Redis error: %v", err))
			log.SetLevel(enum.LogErrorLevel)
		} else {
			logrus.Info("Redis data rolled back...")
			log.SetItem("回滚成功", "Redis data rolled back...")
		}
		log.Save()
		return
	}
	logrus.Infof("update comment complete, total %d comments, %s time elapsed", len(commentList), time.Since(start))
	log.SetItem("完成", fmt.Sprintf("update comment data complete, total %d article(s) involved, %s time elapsed", len(commentList), time.Since(start)))
	log.SetTitle("同步成功")
	log.Save()
}

// RollbackCommentRedis 如果写入 db 失败，将数据回滚到 redis 中
func RollbackCommentRedis(replyMap, likeMap map[uint]int) error {
	// 当前 Redis 中的新增增量
	newReplyMap := redis_comment.GetAllReplyCounts()
	newLikeMap := redis_comment.GetAllLikeCounts()

	// Redis Key
	replyKey := string(redis_comment.CommentReplyCount)
	likeKey := string(redis_comment.CommentLikeCount)

	// 合并保存数据与新增数据 构建最终恢复数据
	mergedReply := mapMergeAndConvert(replyMap, newReplyMap)
	mergedLike := mapMergeAndConvert(likeMap, newLikeMap)

	// 批量回写
	if len(mergedReply) > 0 {
		if err := global.Redis.HMSet(replyKey, mergedReply).Err(); err != nil {
			return fmt.Errorf("HMSet replyCount error: %v", err)
		}
	}
	if len(mergedLike) > 0 {
		if err := global.Redis.HMSet(likeKey, mergedLike).Err(); err != nil {
			return fmt.Errorf("HMSet likeCount error: %v", err)
		}
	}
	return nil
}
