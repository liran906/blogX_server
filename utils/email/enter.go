// Path: ./blogX_server/utils/email_store/enter.go

package email

import (
	"blogX_server/global"
	"encoding/json"
)

type EmailStore struct {
	Email string // 邮箱地址
	Code  string // 验证码
}

func IsValidEmailCode(emailID, code string) (email, msg string, ok bool) {
	// 从 redis 取
	val, err := global.Redis.Get(emailID).Result()
	if err != nil {
		msg = "邮箱ID错误或验证码已过期: " + err.Error()
		return
	}

	// 解析 json
	var storedData EmailStore
	_ = json.Unmarshal([]byte(val), &storedData)

	// 校验 code
	if storedData.Code != code {
		msg = "邮箱验证码错误"
		return
	}
	return storedData.Email, "", true
}
