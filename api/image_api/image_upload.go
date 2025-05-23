// Path: ./blogX_server/api/image_api/image_upload.go

package image_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

func (i *ImageApi) ImageUploadView(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 大小限制
	sizeLimit := global.Config.Upload.ImageSizeLimit // 单位 MB
	if int(fileHeader.Size) > sizeLimit*1024*1024 {
		res.FailWithMsg(fmt.Sprintf("文件大小 %.1fMB, 超过 %dMB限制", float32(fileHeader.Size)/1024/1024, sizeLimit), c)
		return
	}

	// 合法格式
	err = ImageSuffix(fileHeader.Filename)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	filePath := "uploads/" + global.Config.Upload.ImageDir + "/" + fileHeader.Filename
	err = c.SaveUploadedFile(fileHeader, filePath)
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	res.Success("/"+filePath, "成功上传图片 "+fileHeader.Filename, c)
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

func ImageSuffix(filename string) error {
	_list := strings.Split(filename, ".")
	if len(_list) <= 1 {
		return errors.New("非法文件名")
	}

	suffix := _list[len(_list)-1]
	whiteList := global.Config.Upload.ValidImageSuffixes

	if !utils.InList(suffix, whiteList) {
		return errors.New("非法文件格式")
	}
	return nil
}
