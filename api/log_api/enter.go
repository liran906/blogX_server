// Path: ./blogX_server/api/log_api/enter.go

package log_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"fmt"
	"github.com/gin-gonic/gin"
)

type LogApi struct{}

type LogListRequest struct {
	common.PageInfo
	LogType     enum.LogType      `form:"logType"`
	Level       enum.LogLevelType `form:"level"`
	UserID      uint              `form:"userID"`
	Username    string            `form:"username"`
	IP          string            `form:"ip"`
	Address     string            `form:"address"`
	ServiceName string            `form:"serviceName"`
}

// LogListResponse 查询用户时候除了 id，再附带一些其他想展示的信息
type LogListResponse struct {
	models.LogModel
	Username      string `json:"username"`
	UserNickName  string `json:"userNickName"`
	UserAvatarURL string `json:"userAvatarURL"`
}

func (LogApi) LogListView(c *gin.Context) {
	// 分页 查询（精确查询 模糊匹配）
	var req LogListRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	list, count, err := common.ListQuery(
		models.LogModel{ // 精确匹配参数
			LogType:     req.LogType,
			Level:       req.Level,
			UserID:      req.UserID,
			Username:    req.Username,
			IP:          req.IP,
			IPLocation:  req.Address,
			ServiceName: req.ServiceName,
		},
		common.Options{ // 模糊匹配及其他参数
			PageInfo:     req.PageInfo,
			Likes:        []string{"title"},
			Preloads:     []string{"UserModel"},
			Debug:        false,
			DefaultOrder: "id desc",
		},
	)

	// 下面注释的代码，已经封装到 common.ListQuery() 中了
	/*
		// 每一页超过 100 条则按 100 条
		if req.Limit > 100 {
			req.Limit = 100
		}
		// 每一页默认 10，若小于 0 条则按默认
		if req.Limit <= 0 {
			req.Limit = 10
		}
		// 默认第一页开始
		if req.Page <= 0 {
			req.Page = 1
		}

		model := models.LogModel{
			LogType:     req.LogType,
			Level:       req.Level,
			UserID:      req.UserID,
			Username:    req.Username,
			IP:          req.IP,
			IPLocation:     req.IPLocation,
			ServiceName: req.ServiceName,
		}

		// 模糊查询语句生成
		like := global.DB.
			// Model() 指定要查询的数据模型是 LogModel，相当于指定 FROM log_models
			Model(models.LogModel{}).
			// Where() 构建 WHERE 条件：title LIKE '%key%'
			// fmt.Sprintf("%%%s%%", req.Key) 会将 req.Key 包装成 SQL 的模糊匹配格式
			Where("title like ?", fmt.Sprintf("%%%s%%", req.Key))

		// 查询数据库
		var list []models.LogModel
		global.DB.
			Preload("UserModel").     // Preload() 预加载关联数据（比如 FK），这里会同时加载每条日志关联的用户信息
			Model(models.LogModel{}). // 再次指定要查询的数据模型（此处可以省略）
			Where(like).              // 使用上面构建的模糊查询条件
			Where(model).             // 使用传入的 model 参数作为精确匹配条件。非零值字段都会作为 WHERE 条件
			Offset((req.Page - 1) * req.Limit).
			Limit(req.Limit).
			Find(&list) // 执行查询并将结果存入 list 切片中

		// 统计数量
		var count int64
		global.DB.Model(models.LogModel{}).Where(like).Where(model).Count(&count)
	*/

	var _list = make([]LogListResponse, 0)
	for _, logModel := range list {
		_list = append(_list, LogListResponse{
			LogModel:      logModel,
			Username:      logModel.UserModel.Username,
			UserNickName:  logModel.UserModel.Nickname,
			UserAvatarURL: logModel.UserModel.AvatarURL,
		})
	}

	res.SuccessWithList(_list, count, c)
}

func (LogApi) LogReadView(c *gin.Context) {
	var req models.IDRequest
	err := c.ShouldBindUri(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var log models.LogModel
	err = global.DB.Take(&log, req.ID).Error
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	if !log.IsRead {
		global.DB.Model(&log).Update("is_read", true)
	}

	res.Success(log, "读取日志成功", c)
}

func (LogApi) LogRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var removeList []models.LogModel
	global.DB.Find(&removeList, "id in ?", req.IDList)

	var validIDList []uint
	for _, item := range removeList {
		validIDList = append(validIDList, item.ID)
	}

	if len(removeList) > 0 {
		global.DB.Delete(&removeList)

		// 记录删除记录进入日志
		log := log_service.GetActionLog(c)
		log.ShowAll()
		log.SetTitle("日志删除")
		log.SetItem("删除列表: ", removeList)

		msg := fmt.Sprintf("日志删除: 请求 %d 条，成功删除 %d 条，已删除列表: %v", len(req.IDList), len(removeList), validIDList)
		res.SuccessWithMsg(msg, c)
	} else {
		res.FailWithMsg("无匹配日志", c)
	}
}

//删除成功，共计 %d 条
