// Path: ./api/image_api/image_upload.go

package image_api

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/cloud_service/qny_cloud_service"
	"blogX_server/service/log_service"
	"blogX_server/utils/file"
	"blogX_server/utils/hash"
	"blogX_server/utils/jwts"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"os"
	"strings"
)

type respCode int8

const (
	respCodeSuccess respCode = 0
	respCodeFail    respCode = 1
	respCodeDupe    respCode = 2
)

func (rc respCode) String() string {
	switch rc {
	case respCodeSuccess:
		return "上传成功"
	case respCodeFail:
		return "上传失败"
	case respCodeDupe:
		return "上传成功"
	}
	return ""
}

// uploadImages 批量上传图像
func uploadImages(files []*multipart.FileHeader, c *gin.Context) (list []*ImageUploadResponse, count int) {
	// 记录日志
	log := log_service.GetActionLog(c)
	log.ShowRequestHeader()
	log.ShowResponseHeader()
	log.ShowResponse()
	log.SetTitle("上传图片")

	claims := jwts.MustGetClaimsFromGin(c)
	uid := claims.UserID

	for _, fileHeader := range files {
		var rCode respCode
		var msg string
		var hashString string

		byteData, err := readMultiPartFile(fileHeader)
		if err != nil {
			rCode, msg = respCodeFail, err.Error()
		} else {
			// 计算 hashString
			hashString = hash.Md5(byteData)
			// 上传单张
			rCode, msg = uploadImage(byteData, fileHeader.Filename, "uploaded", uid)
		}

		// 日志
		log.SetItem(rCode.String(), msg)
		// 回复列表结构体
		uploadResp := &ImageUploadResponse{
			Filename: fileHeader.Filename,
			Size:     fileHeader.Size,
			Title:    rCode.String(),
			Message:  msg,
			Hash:     hashString,
		}
		// 成功数量（包括dupe）
		if rCode != respCodeFail {
			count++
			uploadResp.Message = "" // 成功不返回msg
		}
		list = append(list, uploadResp)
	}
	return
}

func readMultiPartFile(fileHeader *multipart.FileHeader) (byteData []byte, err error) {
	// 大小限制
	sizeLimit := global.Config.Upload.ImageSizeLimit // 单位 MB
	if int(fileHeader.Size) > sizeLimit*1024*1024 {
		msg := fmt.Sprintf("文件大小 %.1fMB, 超过 %dMB限制", float32(fileHeader.Size)/1024/1024, sizeLimit)
		err = errors.New(msg)
		return
	}

	// 读取文件
	f, err := fileHeader.Open()
	if err != nil {
		return
	}

	// 读取字节流
	byteData, err = io.ReadAll(f)
	if err != nil {
		return
	}
	return
}

func uploadImage(byteData []byte, filename, src string, uid uint) (code respCode, msg string) {
	// 计算 hashString
	hashString := hash.Md5(byteData)

	// 合法格式
	suffix, err := file.ImageSuffix(filename)
	if err != nil {
		return respCodeFail, err.Error()
	}

	// 入库
	model := models.ImageModel{
		Filename: filename,
		Path:     fmt.Sprintf("uploads/%s/%s", global.Config.Upload.ImageDir, hashString+"."+suffix),
		Size:     int64(len(byteData)),
		Hash:     hashString,
		Source:   src,
	}

	// 尝试入库，靠数据库 `hash` 字段 `unique` 去重
	err = global.DB.Create(&model).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			// 找出重复的那个
			var _model models.ImageModel
			global.DB.Take(&_model, "hash = ?", hashString)

			// 首先判断是不是同一个用户上传的
			// 如果不是，则加入关系表中
			relation := models.UserUploadImage{
				UserID:  uid,
				ImageID: _model.ID,
			}
			// 关系表尝试入库
			_err := global.DB.Create(&relation).Error
			if _err != nil {
				if strings.Contains(_err.Error(), "Duplicate entry") {
					msg = fmt.Sprintf("相同用户%d && 相同图片%d", uid, _model.ID)
					logrus.Infof(msg)
					return respCodeFail, msg
				} else {
					return respCodeFail, _err.Error()
				}
			}

			// 成功入库：相同图片，不同用户
			msg := fmt.Sprintf("上传的%s 与已有%s 重复，hash: %s", filename, _model.Filename, hashString)
			logrus.Info(msg)
			return respCodeDupe, msg
		}
		return respCodeFail, err.Error()
	}

	// 存入多对多关系数据库
	err = global.DB.Create(&models.UserUploadImage{
		UserID:  uid,
		ImageID: model.ID,
	}).Error
	if err != nil {
		return respCodeFail, err.Error()
	}

	// DB 部分结束，下面是云存储 or 本地存储 or both
	// 判断是否开启云存储
	if global.Config.Cloud.QNY.Enable {
		url, err := qny_cloud_service.UploadBytes(byteData)
		if err != nil {
			return respCodeFail, err.Error()
		}
		model.Url = url
		// 如果没开启本地存储
		if !global.Config.Cloud.QNY.LocalSave {
			// 更新 url
			global.DB.Model(&model).Update("url", model.Url).Update("path", "")
			msg := "filename: " + filename + " path: " + model.Path
			return respCodeSuccess, msg
		}
	}

	// 把云存储路径更新到 db
	if model.Url != "" {
		global.DB.Model(&model).Update("url", model.Url)
	}

	// 创建文件 这个还是需要 fileHeader
	//err = c.SaveUploadedFile(fileHeader, model.Path)
	//if err != nil {
	//	return respCodeFail, "创建文件失败: " + err.Error()
	//}
	//msg = "filename: " + fileHeader.Filename + " path: " + model.Path
	//return respCodeSuccess, msg

	// 本地存储
	err = os.MkdirAll("uploads/images", 0777)
	if err != nil {
		return respCodeFail, err.Error()
	}
	err = os.WriteFile(model.Path, byteData, 0666)
	if err != nil {
		return respCodeFail, err.Error()
	}
	return respCodeSuccess, fmt.Sprintf("成功上传图片 %s", filename)
}
