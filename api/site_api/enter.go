// Path: ./api/site_api/enter.go

package site_api

import (
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

type SiteApi struct{}

// 每个路由绑定到一个视图（View），也就是对应一个页面

func (s *SiteApi) SiteInfoView(c *gin.Context) {
	// TBD
	log_service.NewLoginSuccess(c, enum.UsernamePasswordLoginType)
	log_service.NewLoginFail(c, enum.UsernamePasswordLoginType, "login fail", "un_test", "pw_test")
	c.JSON(200, gin.H{"message": "test: 站点信息"})
	return
}

type SiteUpdateRequest struct {
	Name string `json:"name"`
}

func (s *SiteApi) SiteUpdateView(c *gin.Context) {
	log := log_service.GetLog(c)

	log.ShowRequestHeader()
	log.ShowRequest()
	log.ShowResponseHeader()
	log.ShowResponse()

	// testing
	log.SetTitle("test_更新站点")
	log.SetItemInfo("test:time", time.Now())
	log.SetItemInfo("test:struct", struct {
		Name string
		Age  int
	}{Name: "test", Age: 12})
	log.SetItemWarn("test:slice", []int{1, 2, 3})
	log.SetItemError("test:string", "hello")
	log.SetItemDebug("test:bool", true)
	log.SetItemInfo("test:int", 1)
	log.SetImage("/xxx/xxx.jpg")
	log.SetLink("test_set_link", "https://www.google.com")
	c.Header("test response header key", "test response header value")

	var req SiteUpdateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Errorf(err.Error())
	}

	//log.Save()

	c.JSON(200, gin.H{"msg": "test: 更新站点信息"})
	return
}
