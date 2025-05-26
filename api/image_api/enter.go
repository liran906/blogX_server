// Path: ./blogX_server/api/image_api/enter.go

package image_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/cloud_service/qny_cloud_service"
	"blogX_server/service/log_service"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"strings"
)

type ImageApi struct{}

type ImageListRequest struct {
	common.PageInfo
	Filename string `form:"filename"`
	Path     string `form:"path"`
	Hash     string `form:"hash"`
	// TBD 根据用户 id 搜索？
}

type ImageListResponse struct {
	models.ImageModel
	WebPath string     `json:"webPath"`
	Users   []UserInfo `json:"users"`
}

type UserInfo struct {
	UserID   uint          `json:"userID"`
	Username string        `json:"username"`
	Role     enum.RoleType `json:"role"`
}

func (ImageApi) ImageUploadView(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		res.FailWithError(err, c)
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

	// 记录日志
	log := log_service.GetActionLog(c)
	log.ShowRequestHeader()
	log.ShowResponseHeader()
	log.ShowResponse()
	log.SetTitle("上传图片")

	var list []*ImageUploadResponse
	var count int
	for _, fileHeader := range files {
		uploadImage(fileHeader, log, &list, &count, c)
	}

	if count == len(files) {
		res.SuccessWithList(list, count, c)
	} else {
		res.WithList(list, len(files), count, c)
	}
}

func (ImageApi) ImageListView(c *gin.Context) {
	var req ImageListRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	_list, count, err := common.ListQuery(models.ImageModel{
		Filename: req.Filename,
		Path:     req.Path,
		Hash:     req.Hash,
	},
		common.Options{
			PageInfo: req.PageInfo,
			Likes:    []string{"filename"},
			Preloads: []string{"Users"},
			Debug:    true, //
		})
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var list = make([]ImageListResponse, 0)
	for _, model := range _list {
		var users []UserInfo
		for _, user := range model.Users {
			users = append(users, UserInfo{
				UserID:   user.ID,
				Username: user.Username,
				Role:     user.Role,
			})
		}

		list = append(list, ImageListResponse{
			ImageModel: model,
			WebPath:    model.WebPath(),
			Users:      users,
		})
	}

	res.SuccessWithList(list, count, c)
}

func (ImageApi) ImageRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var removeList []models.ImageModel
	global.DB.Find(&removeList, "id in ?", req.IDList)

	var validIDList []uint
	for _, item := range removeList {
		validIDList = append(validIDList, item.ID)

		// 有钩子函数就不用显示物理删除文件
		/*
			err := os.Remove(image.Path)
			if err != nil {
				msg := fmt.Sprintf("删除文件失败: %v, 路径: %s", err, image.Path)
				logrus.Warnf(msg)
			}
		*/

		// 如果云端有，也要同步删除，这里只考虑七牛云
		if item.Url != "" {
			// 七牛云
			if strings.Contains(item.Url, global.Config.Cloud.QNY.Uri) {
				err := qny_cloud_service.RemoveFile(item.Url)
				if err != nil {
					msg := fmt.Sprintf("云端删除文件失败: %v, 路径: %s", err, item.Url)
					logrus.Error(msg)
					res.FailWithMsg(msg, c)
					// 这里不返回了
				}
			}
		}
	}
	if len(removeList) > 0 {
		// 使用 Select("Users").Unscoped()
		// 这里的 "Users" 对应的是 ImageModel 中定义的字段名`Users []UserModel`
		err := global.DB.Select("Users").Unscoped().Delete(&removeList).Error
		// 会同时删除：
		// 1. image_models 表中的记录
		// 2. user_upload_images 表中关联的记录
		if err != nil {
			logrus.Errorf("删除图片失败 %s", err.Error())
			res.FailWithError(err, c)
			return
		}

		// 日志
		log := log_service.GetActionLog(c)
		log.ShowAll()
		log.SetTitle("删除图片")
		log.SetItem("删除列表: ", removeList)

		msg := fmt.Sprintf("图片删除: 请求 %d 条，成功删除 %d 条，已删除列表: %v", len(req.IDList), len(removeList), validIDList)
		res.SuccessWithMsg(msg, c)
	} else {
		res.FailWithMsg("无匹配图片", c)
	}
}
