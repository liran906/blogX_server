// Path: ./service/redis_service/redis_jwt/enter.go

package redis_jwt

import (
	"blogX_server/global"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type BlockType int8

const (
	UserBlockType   BlockType = 1 // 用户自己注销登录
	AdminBlockType  BlockType = 2 // 管理员注销登录
	DeviceBlockType BlockType = 3 // 其他设备登录
)

func (b BlockType) String() string {
	return fmt.Sprintf("%d", b)
}

func (b BlockType) Msg() string {
	switch b {
	case UserBlockType:
		return "已注销"
	case AdminBlockType:
		return "禁止登录"
	case DeviceBlockType:
		return "设备下线"
	}
	return "已注销"
}

// ParseBlockType 从字符串类型解析到 BlockType 类型
func ParseBlockType(s string) BlockType {
	switch s {
	case "1":
		return UserBlockType
	case "2":
		return AdminBlockType
	case "3":
		return DeviceBlockType
	default:
		return UserBlockType
	}
}

// BlockJWTToken 将 token 加入黑名单
func BlockJWTToken(token string, value BlockType) {
	// 增加前缀
	key := fmt.Sprintf("jwt_block_%s", token)

	// 解析剩余时间
	cla, err := jwts.ParseToken(token)
	if err != nil || cla == nil {
		logrus.Error("failed to parse token: ", err)
		return
	}
	// 过期时间戳（秒级） - 现在时间戳（秒级）
	remainingTimeInSecond := time.Duration(cla.ExpiresAt-time.Now().Unix()) * time.Second

	// 写入 redis
	_, err = global.Redis.Set(key, value.String(), remainingTimeInSecond).Result()
	if err != nil {
		logrus.Error("failed to set redis: ", err)
		return
	}
	logrus.Infof("token [%s...%s] blocked", token[:4], token[len(token)-4:])
}

// IsBlockedJWTToken checks if a provided JWT token is blocked by querying its status from the Redis database.
func IsBlockedJWTToken(token string) (blockType BlockType, ok bool) {
	claims, err := jwts.ParseToken(token)
	if err != nil {
		return
	}

	// 增加前缀
	key1 := fmt.Sprintf("jwt_block_%s", token)
	key2 := fmt.Sprintf("%dpassword_update", claims.UserID)

	// 查询
	val, err1 := global.Redis.Get(key1).Result()
	pwdTime, err2 := global.Redis.Get(key2).Result()
	if err1 != nil && err2 != nil {
		return
	}

	// 修改密码之后的 token 没问题
	tStamp, _ := strconv.Atoi(pwdTime)
	if err2 == nil && int64(tStamp) < claims.IssuedAt {
		return
	}

	if val == "" {
		blockType = UserBlockType
	} else {
		blockType = ParseBlockType(val)
	}

	return blockType, true
}

func IsBlockedJWTTokenByGin(c *gin.Context) (blockType BlockType, ok bool) {
	token := c.GetHeader("token")
	if token == "" {
		token = c.Query("token")
	}
	return IsBlockedJWTToken(token)
}
