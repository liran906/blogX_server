// Path: ./api/image_api/image_upload.go

package image_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/cloud_service/qny_cloud_service"
	"blogX_server/service/log_service"
	"blogX_server/utils/file"
	"blogX_server/utils/hash"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"strings"
)

type ImageUploadResponse struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Hash     string `json:"hash"`
	Message  string `json:"message"`
	Error    string `json:"error"`
}

// uploadImage 上传单张图像
func uploadImage(fileHeader *multipart.FileHeader, log *log_service.ActionLog, list *[]*ImageUploadResponse, count *int, c *gin.Context) {
	uploadResp := &ImageUploadResponse{
		Filename: fileHeader.Filename,
		Size:     fileHeader.Size,
	}

	// 大小限制
	sizeLimit := global.Config.Upload.ImageSizeLimit // 单位 MB
	if int(fileHeader.Size) > sizeLimit*1024*1024 {
		msg := fmt.Sprintf("文件大小 %.1fMB, 超过 %dMB限制", float32(fileHeader.Size)/1024/1024, sizeLimit)
		log.SetItemError("失败", msg)
		uploadResp.Message, uploadResp.Error = "失败", msg
		*list = append(*list, uploadResp)
		return
	}

	// 合法格式
	suffix, err := file.ImageSuffix(fileHeader.Filename)
	if err != nil {
		log.SetItemError("失败", err.Error())
		uploadResp.Message, uploadResp.Error = "失败", err.Error()
		*list = append(*list, uploadResp)
		return
	}

	// 读取文件
	f, err := fileHeader.Open()
	if err != nil {
		log.SetItemError("失败", err.Error())
		uploadResp.Message, uploadResp.Error = "失败", err.Error()
		*list = append(*list, uploadResp)
		return
	}

	// 读取字节流
	byteData, err := io.ReadAll(f)
	if err != nil {
		log.SetItemError("失败", err.Error())
		uploadResp.Message, uploadResp.Error = "失败", err.Error()
		*list = append(*list, uploadResp)
		return
	}

	// 计算 hashString
	hashString := hash.Md5(byteData)
	uploadResp.Hash = hashString

	// 入库
	model := models.ImageModel{
		Filename: fileHeader.Filename,
		Path:     fmt.Sprintf("uploads/%s/%s", global.Config.Upload.ImageDir, hashString+"."+suffix),
		Size:     fileHeader.Size,
		Hash:     hashString,
	}

	_claims, _ := c.Get("claims")
	claims, ok := _claims.(*jwts.MyClaims)
	if !ok {
		fmt.Println(ok)
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
				UserID:  claims.UserID,
				ImageID: _model.ID,
			}

			_err := global.DB.Create(&relation).Error
			if _err != nil {
				if strings.Contains(_err.Error(), "Duplicate entry") {
					logrus.Infof("相同用户%d %s && 相同图片%d", claims.UserID, claims.Username, model.ID)
					log.SetItemInfo("失败", "相同用户&&相同图片")
					uploadResp.Message, uploadResp.Error = "失败", "请不要上传重复图片"
				} else {
					uploadResp.Message, uploadResp.Error = "失败", _err.Error()
				}
				*list = append(*list, uploadResp)
				return
			}

			// 返回提示
			msg := fmt.Sprintf("上传的%s 与已有%s 重复，hash: %s", fileHeader.Filename, _model.Filename, hashString)
			logrus.Info(msg)
			log.SetItemInfo("成功(Dupe)", msg)
			uploadResp.Message = "成功"
			*count++
		} else {
			uploadResp.Message, uploadResp.Error = "失败", err.Error()
		}
		*list = append(*list, uploadResp)
		return
	}

	// 存入多对多关系数据库
	relation := models.UserUploadImage{
		UserID:  claims.UserID,
		ImageID: model.ID,
	}

	err = global.DB.Create(&relation).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			res.FailWithMsg("请不要上传相同的文件", c)
			logrus.Info("请不要上传相同的文件")
		}
	}

	// 判断是否开启云存储
	if global.Config.Cloud.QNY.Enable {
		url, err := qny_cloud_service.UploadBytes(byteData)
		if err != nil {
			log.SetItemError("失败", err.Error())
			uploadResp.Message, uploadResp.Error = "失败", err.Error()
			*list = append(*list, uploadResp)
			return
		}
		model.Url = url
		// 如果没开启本地存储，直接返回
		if !global.Config.Cloud.QNY.LocalSave {
			// 更新 url
			global.DB.Model(&model).Update("url", model.Url).Update("path", "")

			msg := "filename: " + fileHeader.Filename + " path: " + model.Path
			log.SetItem("成功", msg)
			uploadResp.Message = "成功"
			*list = append(*list, uploadResp)
			*count++
			return
		}
	}

	if model.Url != "" {
		global.DB.Model(&model).Update("url", model.Url)
	}
	// 创建文件
	err = c.SaveUploadedFile(fileHeader, model.Path)
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	msg := "filename: " + fileHeader.Filename + " path: " + model.Path
	log.SetItem("成功", msg)
	uploadResp.Message = "成功"
	*list = append(*list, uploadResp)
	*count++
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
