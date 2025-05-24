// Path: ./blogX_server/api/image_api/enter.go

package image_api

import (
	"blogX_server/models"
)

type ImageApi struct{}

type ImageListResponse struct {
	models.ImageModel
	WebPath string `json:"webPath"`
}

//func (i *ImageApi) ImageListView(c *gin.Context) {
//	common.ListQuery(models.ImageModel{}, common.Options{
//		PageInfo
//		Likes
//		Preloads
//		Where
//		Debug
//		DefaultOrder
//	})
//}
