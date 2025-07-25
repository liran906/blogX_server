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

var template = "<!DOCTYPE html>\n<html lang=\"zh-CN\">\n<head>\n  <meta charset=\"UTF-8\" />\n  <title>%s %s</title>\n</head>\n<body style=\"margin:0;padding:80px 0;background-color:#f5f7fa;font-family:'Segoe UI','Microsoft Yahei',sans-serif;\">\n  <div style=\"max-width:600px;margin:0 auto;background-color:#ffffff;border-radius:8px;box-shadow:0 4px 12px rgba(0,0,0,0.05);overflow:hidden;\">\n    \n    <!-- Header -->\n    <div style=\"background-color:#4f46e5;color:#ffffff;text-align:center;padding:36px 20px;\">\n      <h1 style=\"margin:0;font-size:24px;\">%s</h1>\n    </div>\n    \n    <!-- Content -->\n    <div style=\"padding:30px 28px;color:#333333;\">\n      <h2 style=\"font-size:20px;margin-bottom:16px;color:#333333;\">您好，</h2>\n      <p style=\"font-size:16px;line-height:1.7;margin:12px 0;color:#333333;\">您正在%s <strong>%s</strong>%s，这是我们为您生成的验证码：</p>\n      <p style=\"text-align:center;margin:20px 0;\">\n        <span style=\"display:inline-block;font-size:28px;font-weight:bold;color:#4f46e5;background-color:#f0f2ff;padding:12px 24px;border-radius:6px;\">%s</span>\n      </p>\n      <p style=\"font-size:16px;line-height:1.7;margin:12px 0;\">请在 <strong>%d 分钟内</strong> 输入验证码完成%s。验证码仅在当前%s流程中有效，请勿泄露给他人。</p>\n      <p style=\"font-size:16px;line-height:1.7;margin:12px 0;\">如非本人操作，请忽略本邮件，无需任何处理。</p>\n    </div>\n    \n    <!-- Footer -->\n    <div style=\"font-size:12px;color:#999999;text-align:center;padding:24px;background-color:#fafafa;\">\n      本邮件由系统自动发送，请勿回复。<br>\n      &copy; 2025 %s 版权所有\n    </div>\n    \n  </div>\n</body>\n</html>"

func SendSubscribe(tos []string, category, content string) error {
	subject := fmt.Sprintf("最新%s论文精选推荐", category)
	alias := "Daily Generation"
	return SendEmails(tos, alias, subject, content, true)
}

// SendRegisterCode 注册验证码
func SendRegisterCode(to, code string) error {
	var siteName = global.Config.Site.SiteInfo.EnglishTitle
	var expiry = global.Config.Email.CodeExpiry

	subject := fmt.Sprintf("%s 注册验证码", global.Config.Site.SiteInfo.EnglishTitle)
	action := "注册"
	head := fmt.Sprintf("欢迎加入 %s 🎉", siteName)
	text := fmt.Sprintf(template, siteName, action, head, action, siteName, "", code, expiry, action, action, siteName)
	return SendEmail(to, subject, text, true)
}

// SendResetPasswordCode 重置密码
func SendResetPasswordCode(to, code string, uid uint) error {
	var siteName = global.Config.Site.SiteInfo.EnglishTitle
	var expiry = global.Config.Email.CodeExpiry

	subject := fmt.Sprintf("%s 密码重置", global.Config.Site.SiteInfo.EnglishTitle)
	action := "重置"
	userInfo := fmt.Sprintf(" 的密码（用户id：%d）", uid)
	head := fmt.Sprintf("%s 密码重置", siteName)
	text := fmt.Sprintf(template, siteName, action, head, action, siteName, userInfo, code, expiry, action, action, siteName)
	return SendEmail(to, subject, text, true)
}

// SendVerifyCode 绑定邮箱
func SendVerifyCode(to, code string, uid uint) error {
	var siteName = global.Config.Site.SiteInfo.EnglishTitle
	var expiry = global.Config.Email.CodeExpiry

	subject := fmt.Sprintf("%s 绑定邮箱", global.Config.Site.SiteInfo.EnglishTitle)
	action := "绑定"
	userInfo := fmt.Sprintf(" 的邮箱（用户id：%d）", uid)
	head := fmt.Sprintf("%s 绑定邮箱", siteName)
	text := fmt.Sprintf(template, siteName, action, head, action, siteName, userInfo, code, expiry, action, action, siteName)
	return SendEmail(to, subject, text, true)
}

func SendEmail(to, subject, text string, isHTML bool) error {
	return SendEmails([]string{to}, "", subject, text, isHTML)
}

func SendEmails(tos []string, alias, subject, text string, isHTML bool) error {
	var validEmails []string
	for _, to := range tos {
		if IsValidWithDomain(to) {
			validEmails = append(validEmails, to)
		}
	}
	if len(validEmails) == 0 {
		return errors.New("没有合法邮箱地址")
	}

	em := global.Config.Email
	if alias == "" {
		alias = em.Alias
	}
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", alias, em.SendEmail)
	if len(tos) == 1 {
		e.To = validEmails
	} else {
		e.Bcc = validEmails // 密送
	}
	e.Subject = subject
	if isHTML {
		e.Headers.Add("Content-Type", "text/html; charset=UTF-8")
		e.Headers.Add("MIME-Version", "1.0")
		e.HTML = []byte(text)
	} else {
		e.Text = []byte(text)
	}

	err := e.Send(fmt.Sprintf("%s:%d", em.Domain, em.Port), smtp.PlainAuth("", em.SendEmail, em.AuthCode, em.Domain))
	if err != nil && !strings.Contains(err.Error(), "short response:") {
		logrus.Error("send email error: ", err)
		return err
	}
	return nil
}
