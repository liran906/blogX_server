// Path: ./blogX_server/api/banner_api/enter.go

package banner_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/log_service"
	"fmt"
	"github.com/gin-gonic/gin"
)

type BannerApi struct{}

// crud

type BannerCreateReq struct {
	Activated bool   `json:"activated"`
	URL       string `json:"url" binding:"required"`
	Href      string `json:"href"`
}

func (BannerApi) BannerCreateView(c *gin.Context) {
	var req BannerCreateReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("创建 banner")

	err = global.DB.Create(&req).Error
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	res.SuccessWithMsg("成功添加 banner", c)
}

type BannerListReq struct {
	common.PageInfo
	Activated bool `form:"activated"`
}

func (BannerApi) BannerListView(c *gin.Context) {
	var req BannerListReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	list, count, err := common.ListQuery(
		models.BannerModel{
			Activated: req.Activated,
		},
		common.Options{
			PageInfo: req.PageInfo,
			Likes:    []string{"url"},
			Debug:    false,
		})
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	res.SuccessWithList(list, count, c)
}

func (BannerApi) BannerRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var removeList []models.BannerModel
	global.DB.Find(&removeList, "id in ?", req.IDList)

	var validIDList []uint
	for _, item := range removeList {
		validIDList = append(validIDList, item.ID)
	}

	if len(removeList) > 0 {
		err = global.DB.Delete(&removeList).Error
		if err != nil {
			res.FailWithError(err, c)
			return
		}
		// 日志
		log := log_service.GetActionLog(c)
		log.ShowAll()
		log.SetTitle("删除 banner")
		log.SetItem("删除列表: ", removeList)

		msg := fmt.Sprintf("banner 删除: 请求 %d 条，成功删除 %d 条，已删除列表: %v", len(req.IDList), len(removeList), validIDList)
		res.SuccessWithMsg(msg, c)
	} else {
		res.FailWithMsg("无匹配 banner", c)
	}
}

func (BannerApi) BannerUpdateView(c *gin.Context) {
	// uri 请求 id
	var id models.IDRequest
	err := c.ShouldBindUri(&id)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// json 请求修改内容
	var req BannerCreateReq
	err = c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("banner 更新失败")

	// 验证 id
	var model models.BannerModel
	err = global.DB.Take(&model, id.ID).Error
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	if model.Activated != req.Activated || model.URL != req.URL || model.Href != req.Href {
		err = global.DB.Model(&model).Updates(map[string]any{
			"activated": req.Activated,
			"url":       req.URL,
			"href":      req.Href,
		}).Error
		if err != nil {
			res.FailWithError(err, c)
			return
		}
	}
	log.SetTitle("banner更新成功")
	res.Success(model, fmt.Sprintf("banner[%d] 更新成功", id), c)
}
