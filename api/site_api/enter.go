// Path: ./api/site_api/enter.go

package site_api

import (
	"blogX_server/common/res"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"github.com/gin-gonic/gin"
)

type SiteApi struct{}

// 每个路由绑定到一个视图（View），也就是对应一个页面

func (s *SiteApi) SiteInfoView(c *gin.Context) {
	// TBD
	log_service.NewLoginSuccess(c, enum.UsernamePasswordLoginType)
	log_service.NewLoginFail(c, enum.UsernamePasswordLoginType, "login fail", "un_test", "pw_test")
	res.SuccessWithData("test_xxx", c)
	return
}

type SiteUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}

func (s *SiteApi) SiteUpdateView(c *gin.Context) {
	// 拿取请求中间件中存储的 ActionLog 对象
	log := log_service.GetActionLog(c)

	// 参数校验
	var req SiteUpdateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.SetError("test: ShouldBindJSON error: ", err)
		res.FailWithMsg(err.Error(), c)
		return
	}

	res.SuccessWithMsg("test: update success", c)
	return
}
