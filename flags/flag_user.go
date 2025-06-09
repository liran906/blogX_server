// Path: ./flags/flag_user.go

package flags

import (
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/email_service"
	"blogX_server/utils/pwd"
	"blogX_server/utils/user"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"math/rand"
	"time"
)

type FlagUser struct{}

func (FlagUser) Create() {
	var u models.UserModel

	// 输入角色
	for {
		fmt.Println("选择角色:\n1 - 超级管理员\n2 - 普通用户\n3 - 访客")
		_, err := fmt.Scanln(&u.Role)
		if err != nil || u.Role < 1 || u.Role > 3 {
			fmt.Println("角色输入错误，请输入 1/2/3")
			continue
		}
		break
	}

	// 输入用户名
	for {
		fmt.Println("请输入用户名:")
		_, err := fmt.Scanln(&u.Username)
		if err != nil || len(u.Username) > 32 {
			fmt.Println("用户名输入错误")
			continue
		}
		// 这里用终端创建用户就不判断是否 valid 了，算一个 privilege 吧
		//if msg, ok:= user.IsValidUsername(u.Username); !ok {
		//	fmt.Println(msg)
		//	continue
		//}
		if msg, ok := user.IsAvailableUsername(u.Username); !ok {
			fmt.Println(msg)
			continue
		}
		break
	}

	// 输入密码
	for {
		fmt.Println("请输入密码:")
		pswd, err := terminal.ReadPassword(0)
		if err != nil {
			fmt.Println("读取密码错误: ", err)
			continue
		}
		if !user.IsValidPassword(string(pswd)) {
			fmt.Println("密码不符合要求")
			continue
		}
		fmt.Println("请再次输入密码:")
		rePwd, err := terminal.ReadPassword(0)
		if err != nil {
			fmt.Println("读取密码错误: ", err)
			continue
		}
		if string(pswd) != string(rePwd) {
			fmt.Println("两次密码不一致")
			continue
		}
		u.Password, err = pwd.GenerateFromPassword(string(pswd))
		if err != nil {
			fmt.Println("密码哈希错误")
			continue
		}
		break
	}

	// 邮箱
	for {
		fmt.Println("请输入邮箱:")
		_, err := fmt.Scanln(&u.Email)
		if err != nil || len(u.Username) > 256 || !email_service.IsValidEmail(u.Email) {
			fmt.Println("邮箱输入错误")
			continue
		}
		break
	}

	// 完善信息
	// 注意这里邮箱的逻辑没写
	u.Nickname = fmt.Sprintf("用户%05d", rand.Intn(100000))
	u.RegisterSource = enum.RegisterSourceTerminalType
	u.LastLoginTime = time.Now()

	// 入库
	err := transaction.CreateUserAndUserConfig(u)
	if err != nil {
		fmt.Println("入库失败")
		return
	}

	// 日志
	byteData, _ := json.Marshal(u)
	log := models.LogModel{
		LogType:  enum.ActionLogType,
		Title:    "终端注册用户",
		Content:  string(byteData),
		UserID:   u.ID,
		Username: u.Username,
	}
	err = global.DB.Create(&log).Error
	if err != nil {
		logrus.Error("日志入库失败")
	}
	fmt.Println("创建成功")
}
