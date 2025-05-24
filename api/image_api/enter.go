// Path: ./blogX_server/api/image_api/enter.go

package image_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"github.com/gin-gonic/gin"
)

type ImageApi struct{}

type ImageListRequest struct {
	common.PageInfo
	Filename string `form:"filename"`
	Path     string `form:"path"`
	Hash     string `form:"hash"`
	//UserID      uint   `form:"userID"`
	//Username    string `form:"username"`
	//IP          string `form:"ip"`
	//Address     string `form:"address"`
	//ServiceName string `form:"serviceName"`
	//UA          string `form:"ua"`
}

type ImageListResponse struct {
	models.ImageModel
	Users []UserInfo `json:"users"`
}

type UserInfo struct {
	UserID   uint          `json:"userID"`
	Username string        `json:"username"`
	Role     enum.RoleType `json:"role"`
}

func (i *ImageApi) ImageUploadView(c *gin.Context) {
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

func (i *ImageApi) ImageListView(c *gin.Context) {
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
		model.Path = model.WebPath()
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
			Users:      users,
		})
	}

	res.SuccessWithList(list, count, c)
}
