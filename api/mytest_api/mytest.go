// Path: ./api/mytest_api/mytest.go

package mytest_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"github.com/gin-gonic/gin"
)

type MyTestApi struct{}

func (MyTestApi) MyTestView(c *gin.Context) {
	res.Success(global.Test, "Key 值", c)
	redis := global.Redis
	val, err := redis.Get(global.Test).Result()
	res.Success(err, val, c)
}
