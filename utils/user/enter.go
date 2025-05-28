// Path: ./blogX_server/utils/user/enter.go

package user

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/utils"
	"regexp"
)

// IsValidUsername validates the given username to ensure it contains only alphanumeric characters and underscores.
// It also checks against a blacklist of invalid usernames. Returns a message and a boolean indicating validity.
func IsValidUsername(username string) (msg string, ok bool) {
	var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(username) {
		return "用户名含有非法字符", false
	}
	blackList := global.Config.Filter.InvalidUsername
	if utils.InStringList(username, blackList) {
		return "非法用户名", false
	}
	return "", true
}

// IsAvailableUsername checks if a given username is available for registration by querying the database.
// It returns a message and a boolean indicating whether the username is available.
func IsAvailableUsername(username string) (msg string, ok bool) {
	var user models.UserModel
	err := global.DB.Take(&user, "username = ?", username).Error
	if err == nil {
		return "用户名已存在", false
	} else if err.Error() != "record not found" {
		return "读取数据库错误", false
	}
	return "", true
}

// IsValidPassword checks if the given password meets the validity criteria: at least 8 characters, containing letters and digits.
func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	// 至少包含一个字母
	hasLetter, _ := regexp.MatchString(`[a-zA-Z]`, password)
	// 至少包含一个数字
	hasDigit, _ := regexp.MatchString(`[0-9]`, password)

	return hasLetter && hasDigit
}
