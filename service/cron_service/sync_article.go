// Path: ./service/cron_service/sync_article.go

package cron_service

import (
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/service/redis_service/redis_article"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

func SyncArticle() {
	// 记录时间
	start := time.Now()

	log := log_service.NewRuntimeLog("同步文章数据", log_service.RuntimeDeltaDay)
	log.SetItem("开始时间", start.Format("2006-01-02 15:04:05"))

	// 从 redis 中读取数据
	readMap := redis_article.GetAllReadCounts()
	likeMap := redis_article.GetAllLikeCounts()
	collectMap := redis_article.GetAllCollectCounts()
	commentMap := redis_article.GetAllCommentCounts()

	// redis 归零（同步期间如果有增量继续记录，明天再统计）
	redis_article.Clear()

	// 创建一个字典记录有改变的 aid
	var activeArticles = make(map[uint]struct{})
	maps := map[string]map[uint]int{"read": readMap, "like": likeMap, "collect": collectMap, "comment": commentMap}
	for _, m := range maps {
		for aid := range m { // 只要 key
			activeArticles[aid] = struct{}{}
		}
	}

	if len(activeArticles) == 0 {
		logrus.Info("no active article to sync")
		log.SetTitle("无新数据")
		log.Save()
		return
	}

	log.SetTitle("同步失败")

	// 从 DB 中取出本次有修改的文章
	var articleList []models.ArticleModel
	err := global.DB.Where("id IN ?", mapKeys(activeArticles)).Find(&articleList).Error
	if err != nil {
		logrus.Errorf("get article list error: %v", err)
		log.SetItemError("查询失败", fmt.Sprintf("get article list error: %v", err))
		log.SetLevel(enum.LogErrorLevel)
		log.Save()
		return
	}

	// 事务中遍历comment，修改数据
	err = transaction.SyncArticleTx(articleList, maps)
	if err != nil {
		logrus.Errorf("sync article error: %v", err)
		log.SetItemWarn("事务失败", fmt.Sprintf("sync article error: %v", err))
		log.SetLevel(enum.LogWarnLevel)
		if err = RollbackArticleRedis(readMap, likeMap, collectMap, commentMap); err != nil {
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
	logrus.Infof("update article data complete, total %d article(s) involved, %s time elapsed", len(articleList), time.Since(start))
	log.SetItem("完成", fmt.Sprintf("update article data complete, total %d article(s) involved, %s time elapsed", len(articleList), time.Since(start)))
	log.SetTitle("同步成功")
	log.Save()
}

// RollbackArticleRedis 如果写入 db 失败，将数据回滚到 redis 中
func RollbackArticleRedis(readMap, likeMap, collectMap, commentMap map[uint]int) error {
	// 当前 Redis 中的新增增量
	newReadMap := redis_article.GetAllReadCounts()
	newLikeMap := redis_article.GetAllLikeCounts()
	newCollectMap := redis_article.GetAllCollectCounts()
	newCommentMap := redis_article.GetAllCommentCounts()

	// Redis Key
	readKey := string(redis_article.ArticleReadCount)
	likeKey := string(redis_article.ArticleLikeCount)
	collectKey := string(redis_article.ArticleCollectCount)
	commentKey := string(redis_article.ArticleCommentCount)

	// 合并保存数据与新增数据 构建最终恢复数据
	mergedRead := mapMergeAndConvert(readMap, newReadMap)
	mergedLike := mapMergeAndConvert(likeMap, newLikeMap)
	mergedCollect := mapMergeAndConvert(collectMap, newCollectMap)
	mergedComment := mapMergeAndConvert(commentMap, newCommentMap)

	// 批量回写
	if len(mergedRead) > 0 {
		if err := global.Redis.HMSet(readKey, mergedRead).Err(); err != nil {
			return fmt.Errorf("HMSet readCount error: %v", err)
		}
	}
	if len(mergedLike) > 0 {
		if err := global.Redis.HMSet(likeKey, mergedLike).Err(); err != nil {
			return fmt.Errorf("HMSet likeCount error: %v", err)
		}
	}
	if len(mergedCollect) > 0 {
		if err := global.Redis.HMSet(collectKey, mergedCollect).Err(); err != nil {
			return fmt.Errorf("HMSet collectCount error: %v", err)
		}
	}
	if len(mergedComment) > 0 {
		if err := global.Redis.HMSet(commentKey, mergedComment).Err(); err != nil {
			return fmt.Errorf("HMSet commentCount error: %v", err)
		}
	}
	return nil
}
