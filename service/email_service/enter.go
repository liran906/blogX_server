// Path: ./service/email_service/enter.go

package email_service

import (
	"blogX_server/global"
	"errors"
	"fmt"
	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"
	"net/smtp"
	"strings"
)

// SendRegisterCode 注册验证码
func SendRegisterCode(to, code string) error {
	subject := fmt.Sprintf("%s 注册验证码", global.Config.Site.SiteInfo.EnglishTitle)
	text := fmt.Sprintf("您正在注册 %s 会员，验证码: %s，%d分钟内有效", global.Config.Site.SiteInfo.EnglishTitle, code, global.Config.Email.CodeExpiry)
	return SendEmail(to, subject, text)
}

// SendResetPasswordCode 重置密码
func SendResetPasswordCode(to, code string, uid uint) error {
	subject := fmt.Sprintf("%s 密码重置", global.Config.Site.SiteInfo.EnglishTitle)
	text := fmt.Sprintf("您正在重置 %s 密码，会员id: %d，验证码: %s，%d分钟内有效", global.Config.Site.SiteInfo.EnglishTitle, uid, code, global.Config.Email.CodeExpiry)
	return SendEmail(to, subject, text)
}

// SendVerifyCode 绑定邮箱
func SendVerifyCode(to, code string, uid uint) error {
	subject := fmt.Sprintf("%s 绑定邮箱", global.Config.Site.SiteInfo.EnglishTitle)
	text := fmt.Sprintf("您正在绑定 %s 邮箱，会员id: %d，验证码: %s，%d分钟内有效", global.Config.Site.SiteInfo.EnglishTitle, uid, code, global.Config.Email.CodeExpiry)
	return SendEmail(to, subject, text)
}

func SendEmail(to, subject, text string) error {
	if !IsValidWithDomain(to) {
		return errors.New("非法邮箱地址")
	}
	em := global.Config.Email
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", em.Alias, em.SendEmail)
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(text)
	err := e.Send(fmt.Sprintf("%s:%d", em.Domain, em.Port), smtp.PlainAuth("", em.SendEmail, em.AuthCode, em.Domain))
	if err != nil && !strings.Contains(err.Error(), "short response:") {
		logrus.Error("send email error: ", err)
		return err
	}
	return nil
}
