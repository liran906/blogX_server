// Path: ./api/user_api/register_email.go

package user_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"blogX_server/utils/pwd"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"time"
)

type RegisterEmailReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (UserApi) RegisterEmailView(c *gin.Context) {
	req := c.MustGet("bindReq").(RegisterEmailReq)
	// 注释的逻辑用中间件实现了
	/*
		// 判断是合法 username
		// 后面可以封装
		var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
		if !usernameRegex.MatchString(req.Username) {
			res.FailWithMsg("非法用户名", c)
			return
		}
		blackList := global.Config.Filter.InvalidUsername
		if utils.InStringList(req.Username, blackList) {
			res.FailWithMsg("用户名含有非法字符", c)
			return
		}

		// 判断 username 是否重复
		var user models.UserModel
		err = global.DB.Take(&user, "username = ?", req.Username).Error
		if err == nil {
			res.FailWithMsg("用户名已存在", c)
			return
		} else if err.Error() != "record not found" {
			// 读取数据库错误
			res.FailWithError(err, c)
			return
		}

		// 从 redis 取
		val, err := global.Redis.Get(req.EmailID).Result()
		if err != nil {
			res.FailWithMsg("邮箱ID错误或验证码已过期: "+err.Error(), c)
			return
		}

		// 解析 JSON
		var storedData email.EmailStore
		_ = json.Unmarshal([]byte(val), &storedData)

		if storedData.Code != req.EmailCode {
			res.FailWithMsg("邮箱验证码错误", c)
			return
		}
	*/

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowRequestHeader()
	log.ShowResponseHeader()
	log.ShowResponse()
	log.SetTitle("邮箱注册")

	// 从中间件存储的 email 字段中拿取 email 地址
	emailAddr, ok := c.Get("email")
	if !ok {
		res.FailWithMsg("系统读取邮箱地址错误", c)
		return
	}

	// 密码加盐并哈希
	hashPwd, err := pwd.GenerateFromPassword(req.Password)
	if err != nil {
		res.FailWithMsg("密码设置错误: "+err.Error(), c)
		return
	}
	// 创建
	user := models.UserModel{
		Username:       req.Username,
		Nickname:       fmt.Sprintf("用户%05d", rand.Intn(100000)),
		Email:          emailAddr.(string),
		Password:       hashPwd,
		RegisterSource: enum.RegisterSourceEmailType,
		Role:           enum.UserRoleType,
		LastLoginIP:    c.ClientIP(),
		LastLoginTime:  time.Now(),
	}
	userConf := models.UserConfigModel{
		Tags:               []string{},
		ThemeID:            1, // 默认主题
		DisplayCollections: true,
		DisplayFans:        true,
		DisplayFollowing:   true,
	}
	err = common.CreateUserAndUserConfig(user, userConf)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 颁发 token
	token, err := jwts.GenerateToken(jwts.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		res.FailWithMsg("邮箱登录失败: "+err.Error(), c)
		return
	}

	eid := c.MustGet("emailID").(string)
	global.Redis.Del(eid)

	// 返回 token 与成功信息
	res.Success(token, "成功创建用户", c)
}
