// Path: ./blogX_server/api/image_api/image_upload.go

package image_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/log_service"
	"blogX_server/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
)

func (i *ImageApi) ImageUploadView(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	log := log_service.GetActionLog(c)
	log.ShowRequestHeader()
	log.ShowResponseHeader()
	log.ShowResponse()
	log.SetTitle("上传图片")

	// 大小限制
	sizeLimit := global.Config.Upload.ImageSizeLimit // 单位 MB
	if int(fileHeader.Size) > sizeLimit*1024*1024 {
		msg := fmt.Sprintf("文件大小 %.1fMB, 超过 %dMB限制", float32(fileHeader.Size)/1024/1024, sizeLimit)
		log.SetItemError("文件过大", msg)
		res.FailWithMsg(msg, c)
		return
	}

	// 合法格式
	suffix, err := ImageSuffix(fileHeader.Filename)
	if err != nil {
		log.SetItemError("非法格式", err.Error())
		res.FailWithError(err, c)
		return
	}

	// 文件 hash
	file, err := fileHeader.Open()
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	byteData, _ := io.ReadAll(file)
	hash := utils.Md5(byteData)

	// 入库
	model := models.ImageModel{
		Filename: fileHeader.Filename,
		Path:     fmt.Sprintf("uploads/%s/%s", global.Config.Upload.ImageDir, hash+"."+suffix),
		Size:     fileHeader.Size,
		Hash:     hash,
	}

	err = global.DB.Create(&model).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			// 找出重复的那个
			var _model models.ImageModel
			global.DB.Take(&_model, "hash = ?", hash)

			// 返回提示
			msg := fmt.Sprintf("用户 上传的图片%s 与已有图片%s 重复，hash: %s", fileHeader.Filename, _model.Filename, hash)
			logrus.Info(msg)
			log.SetItemInfo("重复文件", msg)
			res.Success(model.WebPath(), "成功上传图片 "+fileHeader.Filename, c)
		} else {
			res.FailWithError(err, c)
		}
		return
	}

	// 创建文件
	err = c.SaveUploadedFile(fileHeader, model.Path)
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	res.Success(model.WebPath(), "成功上传图片 "+fileHeader.Filename, c)
	/*
		c.SaveUploadedFile 自动完成下面所有这些：
		file, err := fileHeader.Open()
		if err != nil {
			res.FailWithError(err, c)
			return
		}
		byteData, err := io.ReadAll(file)
		if err != nil {
			res.FailWithError(err, c)
			return
		}
		err = os.MkdirAll("uploads/images", 0777)
		if err != nil {
			res.FailWithError(err, c)
			return
		}
		err = os.WriteFile("uploads/images/"+fileHeader.Filename, byteData, 0666)
		if err != nil {
			res.FailWithError(err, c)
			return
		}
		res.SuccessWithMsg(fmt.Sprintf("成功上传图片 %sizeLimit", fileHeader.Filename), c)
	*/
}

func ImageSuffix(filename string) (suffix string, err error) {
	_list := strings.Split(filename, ".")
	if len(_list) <= 1 {
		err = errors.New("非法文件名")
		return
	}

	suffix = _list[len(_list)-1]
	whiteList := global.Config.Upload.ValidImageSuffixes

	if !utils.InList(suffix, whiteList) {
		err = errors.New("非法文件格式")
		return
	}
	return
}
