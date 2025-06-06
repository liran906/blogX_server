// Path: ./service/cron_service/sync_article.go

package cron_service

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/redis_service/redis_article"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

func SyncArticle() {
	// 记录时间
	start := time.Now()

	// 从 redis 中读取数据
	readMap := redis_article.GetAllReadCounts()
	likeMap := redis_article.GetAllLikeCounts()
	collectMap := redis_article.GetAllCollectCounts()
	commentMap := redis_article.GetAllCommentCounts()

	// redis 归零（同步期间如果有增量继续记录，明天再统计）
	redis_article.Clear()

	// 创建一个字典记录有改变的 aid
	var activeArticles = make(map[uint]struct{})
	maps := []map[uint]int{readMap, likeMap, collectMap, commentMap}
	for _, m := range maps {
		for aid := range m { // 只要 key
			activeArticles[aid] = struct{}{}
		}
	}

	if len(activeArticles) == 0 {
		logrus.Info("no active article to sync")
		return
	}

	// 从 DB 中取出对应的文章
	var articleList []models.ArticleModel
	err := global.DB.Where("id IN ?", mapKeys(activeArticles)).Find(&articleList).Error
	if err != nil {
		logrus.Errorf("get article list error: %v", err)
		return
	}

	// 遍历文章，修改数据
	count := 0
	for _, article := range articleList {

		// 每篇文章提取出更新数据
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

		// 如果有全为 0 的情况，上面 activeArticles 是无法筛选出来的，所以这里再筛一次
		if len(updateMap) == 0 {
			continue
		}

		// 写入数据库
		err = global.DB.Model(&article).Updates(updateMap).Error
		if err != nil {
			logrus.Errorf("update article[%d] error: %v", article.ID, err)
			continue
		}
		logrus.Infof("update article[%d]", article.ID)
		count++
	}
	logrus.Infof("update articles complete, total %d articles, %d success, %s time elapsed", len(articleList), count, time.Since(start))
}

func mapKeys(m map[uint]struct{}) []uint {
	keys := make([]uint, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
