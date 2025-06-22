// Path: ./api/captcha_api/enter.go

package captcha_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"github.com/sirupsen/logrus"
	"image/color"
)

type CaptchaApi struct{}

type CaptchaResponse struct {
	CaptchaID string `json:"captchaID"`
	Captcha   string `json:"captcha"`
}

func (CaptchaApi) CaptchaView(c *gin.Context) {
	var driver base64Captcha.Driver
	var driverString base64Captcha.DriverString

	// 配置验证码信息
	captchaConfig := base64Captcha.DriverString{
		Height:     60,
		Width:      200,
		NoiseCount: 1,
		// 控制显示在验证码图片中的线条的选项
		// 1: 直线; 2: 曲线; 4: 点线; 8: 虚线; 16: 中空直线; 32: 中空曲线
		ShowLineOptions: 1 | 2,
		Length:          4,
		Source:          "1234567890ABCDEFGHJKLMNPQRSTUVWXYZ",
		BgColor: &color.RGBA{
			// 背景颜色
			R: 255, // 红色
			G: 255, // 绿色
			B: 255, // 蓝色
			A: 100, // 透明度
		},
		// 字体文件
		//Fonts: []string{"wqy-microhei.ttc"},
	}

	driverString = captchaConfig
	driver = driverString.ConvertFonts()
	captcha := base64Captcha.NewCaptcha(driver, global.CaptchaStore)
	// lid是生成的验证码的唯一标识符 b64是验证码图片的Base64编码字符串
	lid, b64, _, err := captcha.Generate()
	if err != nil {
		logrus.Error("captcha generate error: ", err)
		res.FailWithMsg("captcha generate error: "+err.Error(), c)
		return
	}

	res.SuccessWithData(CaptchaResponse{
		CaptchaID: lid,
		Captcha:   b64,
	}, c)
	return
}
