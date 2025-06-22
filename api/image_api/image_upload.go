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
	"os"
	"strings"
)

type ImageUploadResponse struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Hash     string `json:"hash"`
	Title    string `json:"title"`
	Message  string `json:"message"`
}

// ImageUploadView 上传单张，返回 url（优先云端）
func (ImageApi) ImageUploadView(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	// 文件大小判断
	s := global.Config.Upload.ImageSizeLimit
	if fileHeader.Size > int64(s*1024*1024) {
		res.FailWithMsg(fmt.Sprintf("文件大小大于%dMB", s), c)
		return
	}

	// 记录日志
	log := log_service.GetActionLog(c)
	log.ShowRequestHeader()
	log.ShowResponseHeader()
	log.ShowResponse()
	log.SetTitle("上传图片")

	claims := jwts.MustGetClaimsFromRequest(c)
	uid := claims.UserID

	var msg string

	byteData, err := readMultiPartFile(fileHeader)
	if err != nil {
		res.Fail(err, "读取文件失败", c)
		return
	}

	// 计算 hashString
	hashString := hash.Md5(byteData)
	filename := fileHeader.Filename

	// 格式校验
	suffix, err := file.ImageSuffix(filename)
	if err != nil {
		res.Fail(err, "非法文件格式", c)
		return
	}

	// 入库
	imgModel := models.ImageModel{
		Filename: filename,
		Path:     fmt.Sprintf("uploads/%s/%s", global.Config.Upload.ImageDir, hashString+"."+suffix),
		Size:     int64(len(byteData)),
		Hash:     hashString,
		Source:   "uploaded",
	}

	// 尝试入库，靠数据库 `hash` 字段 `unique` 去重
	err = global.DB.Create(&imgModel).Error
	if err != nil {
		var dupeImgModel models.ImageModel
		if strings.Contains(err.Error(), "Duplicate entry") {
			// 找出重复的那个

			global.DB.Take(&dupeImgModel, "hash = ?", hashString)

			// 首先判断是不是同一个用户上传的
			// 如果不是，则加入关系表中
			relation := models.UserUploadImage{
				UserID:  uid,
				ImageID: dupeImgModel.ID,
			}
			// 关系表尝试入库
			_err := global.DB.Create(&relation).Error
			if _err != nil {
				if strings.Contains(_err.Error(), "Duplicate entry") {
					msg = fmt.Sprintf("相同用户%d && 相同图片%d", uid, dupeImgModel.ID)
					logrus.Infof(msg)
					res.Fail(err, msg, c)
					return
				}
				res.FailWithError(err, c)
				return
			}
		}
		// 成功入库：相同图片，不同用户
		msg := fmt.Sprintf("上传的%s 与已有%s 重复，hash: %s", filename, dupeImgModel.Filename, hashString)
		log.SetItemTrace("重复", msg)
		res.Success(dupeImgModel.WebPath(), "上传成功", c)
		return
	}

	// 存入多对多关系数据库
	err = global.DB.Create(&models.UserUploadImage{
		UserID:  uid,
		ImageID: imgModel.ID,
	}).Error
	if err != nil {
		res.Fail(err, "上传失败", c)
		return
	}

	// DB 部分结束，下面是云存储 or 本地存储 or both
	// 判断是否开启云存储
	if global.Config.Cloud.QNY.Enable {
		url, err := qny_cloud_service.UploadBytes(byteData)
		if err != nil {
			res.Fail(err, "上传云端失败", c)
			return
		}
		imgModel.Url = url
		// 如果没开启本地存储
		if !global.Config.Cloud.QNY.LocalSave {
			// 更新 url
			global.DB.Model(&imgModel).Update("url", imgModel.Url).Update("path", "")
			res.Success(imgModel.Url, msg, c)
			return
		}
	}

	// 把云存储路径更新到 db
	if imgModel.Url != "" {
		global.DB.Model(&imgModel).Update("url", imgModel.Url)
	}

	// 本地存储
	err = os.MkdirAll("uploads/images", 0777)
	if err != nil {
		res.Fail(err, "上传服务器失败", c)
		return
	}
	err = os.WriteFile(imgModel.Path, byteData, 0666)
	if err != nil {
		res.Fail(err, "上传服务器失败", c)
		return
	}
	if imgModel.Url != "" { // 优先云端
		res.Success(imgModel.Url, msg, c)
	} else {
		res.Success(imgModel.WebPath(), msg, c)
	}
}

// ImageBatchUploadView 批量上传
func (ImageApi) ImageBatchUploadView(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		res.Fail(err, "文件读取错误", c)
		return
	}

	files := form.File["file"] // 前端上传时使用 file 作为 key
	if len(files) == 0 {
		res.FailWithMsg("没有上传任何文件", c)
		return
	}
	if len(files) > 10 {
		res.FailWithMsg("一次上传不能超过10张", c)
		return
	}

	list, count := uploadImages(files, c)

	if count == len(files) {
		res.SuccessWithList(list, count, c)
	} else {
		res.WithList(list, len(files), count, c)
	}
}
