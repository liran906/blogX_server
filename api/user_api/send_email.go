// Path: ./api/user_api/send_email.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/email_service"
	"blogX_server/utils/email"
	"blogX_server/utils/jwts"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"gorm.io/gorm"
	"time"
)

type SendEmailReq struct {
	Type  int8   `json:"type" binding:"oneof= 1 2 3"` // 1注册 2密码重置 3绑定邮箱
	Email string `json:"email" binding:"required"`
}

type SendEmailResponse struct {
	EmailID string `json:"emailID"`
}

func (UserApi) SendEmailView(c *gin.Context) {
	req := c.MustGet("bindReq").(SendEmailReq)

	// 站点未启用邮箱注册
	if req.Type == 1 && !global.Config.Site.Login.EmailRegister {
		res.FailWithMsg("站点未启用邮箱注册", c)
		return
	}

	// 验证地址合法
	if !email_service.IsValidWithDomain(req.Email) {
		res.FailWithMsg("非法邮箱地址", c)
		return
	}

	// 这里借用一下验证码现成的方法，生成存储
	code := base64Captcha.RandText(4, "1234567890") // 生成随机验证码
	key := base64Captcha.RandomId()                 // 生成不重复 id; tbd：在这里或者其他地方最好保存下 type。也是边界情况之一。
	// 获取邮箱对应的用户信息
	var user models.UserModel
	var userID uint = 0
	var err error

	switch req.Type {
	case 1:
		err = global.DB.Take(&user, "email = ?", req.Email).Error
		// 检查是否已注册
		if err == nil {
			res.FailWithMsg("该邮箱已被注册", c)
			return
		}
		// 发送
		err = email_service.SendRegisterCode(req.Email, code)
	case 2:
		err = global.DB.Take(&user, "email = ?", req.Email).Error
		// 检查是否已注册
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				res.Fail(err, "该邮箱未注册", c)
			} else {
				res.FailWithError(err, c)
			}
			return
		}
		// 发送
		err = email_service.SendResetPasswordCode(req.Email, code, user.ID)
		userID = user.ID
	case 3:
		err = global.DB.Take(&user, "email = ?", req.Email).Error
		// 检查是否已注册
		if err == nil {
			res.FailWithMsg("该邮箱已被注册", c)
			return
		}
		err = nil

		// 取用户信息
		claims, err := jwts.ParseTokenFromRequest(c)
		if err != nil {
			res.Fail(err, "请登录", c)
			return
		}
		userID = claims.UserID
		// 发送
		err = email_service.SendVerifyCode(req.Email, code, userID)
	}
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 存入 redis
	emStruct := email.EmailStore{Email: req.Email, Code: code, UserID: userID}
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

	res.Success(SendEmailResponse{EmailID: key}, "成功发送邮件", c)
}
