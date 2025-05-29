// Path: ./blogX_server/api/user_api/send_email.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/email_service"
	"blogX_server/utils/email"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"strconv"
	"strings"
	"time"
)

type SendEmailRequest struct {
	Type  int8   `json:"type" binding:"oneof= 1 2 3"` // 1注册 2密码重置 3绑定邮箱
	Email string `json:"email" binding:"required"`
}

type SendEmailResponse struct {
	EmailID string `json:"emailID"`
}

func (UserApi) SendEmailView(c *gin.Context) {
	var req SendEmailRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 站点未启用邮箱注册
	if req.Type == 1 && !global.Config.Site.Login.EmailRegister {
		res.FailWithMsg("站点未启用邮箱注册", c)
		c.Abort()
		return
	}

	// 验证地址合法
	if !email_service.IsValidWithDomain(req.Email) {
		res.FailWithMsg("非法邮箱地址", c)
		return
	}

	// 这里借用一下验证码现成的方法，生成存储
	code := base64Captcha.RandText(4, "1234567890")               // 生成随机验证码
	key := base64Captcha.RandomId() + strconv.Itoa(int(req.Type)) // 生成不重复 id

	// 获取邮箱对应的用户信息
	var user models.UserModel
	err = global.DB.Take(&user, "email = ?", req.Email).Error
	switch req.Type {
	case 1:
		// 检查是否已注册
		if err == nil {
			res.FailWithMsg("该邮箱已被注册", c)
			return
		}
		// 发送
		err = email_service.SendRegisterCode(req.Email, code)
	case 2:
		// 检查是否已注册
		if err != nil {
			if strings.Contains(err.Error(), "record not found") {
				res.FailWithMsg("该邮箱未注册", c)
			} else {
				res.FailWithError(err, c)
			}
			return
		}
		// 发送
		err = email_service.SendResetPasswordCode(req.Email, code, user.ID)
	case 3:
		// 检查是否已注册
		if err == nil {
			res.FailWithMsg("该邮箱已被注册", c)
			return
		}
		// 发送
		err = email_service.SendVerifyCode(req.Email, code, user.ID)
	}
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 存入 redis
	emStruct := email.EmailStore{Email: req.Email, Code: code}
	jsonStr, err := json.Marshal(emStruct)
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	_, err = global.Redis.Set(
		key,
		jsonStr,
		(time.Duration(global.Config.Email.CodeExpiry))*time.Minute,
	).Result()
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	res.Success(key, "成功发送邮件", c)
}
