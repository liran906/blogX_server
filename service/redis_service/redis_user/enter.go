// Path: ./service/redis_service/redis_user/enter.go

package redis_user

import (
	"blogX_server/global"
	"blogX_server/models"
	"github.com/sirupsen/logrus"
	"strconv"
)

const key = "homepage_visit_count"

// IncreaseHPVCount 用户主页流量+1
func IncreaseHPVCount(userID uint) {
	global.Redis.HIncrBy(key, strconv.Itoa(int(userID)), 1)
}

// 这里功能不是很重要，就不返回错误了

func GetHPVCount(userID uint) int {
	count, _ := global.Redis.HGet(key, strconv.Itoa(int(userID))).Int()
	return count
}

func UpdateHPVCount(uc *models.UserConfigModel) {
	uc.HomepageVisitCount += GetHPVCount(uc.UserID)
}

func ClearUserHPVCount(userID uint) {
	global.Redis.HDel(key, strconv.Itoa(int(userID)))
}

func Clear() {
	err := global.Redis.Del(key).Err()
	if err != nil {
		logrus.Errorf("Failed to clear user redis cache: %v", err)
	}
}
