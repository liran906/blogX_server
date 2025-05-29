// Path: ./blogX_server/api/user_api/user_login_list.go

package user_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
	"time"
)

type UserLoginListRequest struct {
	common.PageInfo
	UserID     uint   `form:"userID"`
	IP         string `form:"ip"`
	IPLocation string `form:"ipLocation"`
	StartTime  string `form:"startTime"` // format "2006-01-02 15:04:05"
	EndTime    string `form:"endTime"`
}

type UserLoginListResponse struct {
	ID           uint           `json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	Title        string         `json:"title"`
	IP           string         `json:"ip"`
	IPLocation   string         `json:"ipLocation"`
	LoginType    enum.LoginType `json:"loginType"`
	UA           string         `json:"ua,omitempty"`
	UserID       uint           `json:"userID"`
	Username     string         `json:"username,omitempty"`
	UserNickName string         `json:"userNickName,omitempty"`
}

func (UserApi) UserLoginListView(c *gin.Context) {
	var req UserLoginListRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	claims, ok := jwts.GetClaimsFromGin(c)
	if !ok {
		res.FailWithMsg("获取信息失败，请重新登录", c)
		return
	}

	// 非管理员只能看自己的记录, 且不关联用户信息
	var preloads = []string{"UserModel"}
	if claims.Role != enum.AdminRoleType {
		req.UserID = claims.UserID
		preloads = []string{}
	}
	if req.UserID == 0 {
		req.UserID = claims.UserID
	}

	// 解析时间戳并查询
	query := global.DB.Where("")
	if req.StartTime != "" {
		_, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
		if err != nil {
			res.FailWithMsg("开始时间格式错误", c)
			return
		}
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		_, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
		if err != nil {
			res.FailWithMsg("结束时间格式错误", c)
			return
		}
		query = query.Where("created_at <= ?", req.EndTime)
	}

	_list, count, err := common.ListQuery(
		models.LogModel{ // 精确匹配参数
			LogType:    enum.LoginLogType,
			UserID:     req.UserID,
			IP:         req.IP,
			IPLocation: req.IPLocation,
		},
		common.Options{ // 模糊匹配及其他参数
			PageInfo:     req.PageInfo,
			Likes:        []string{"title"},
			Preloads:     preloads,
			Where:        query,
			Debug:        false,
			DefaultOrder: "id desc",
		},
	)
	if err != nil {
		res.FailWithMsg("查询失败: "+err.Error(), c)
		return
	}

	// 注意这里如果 make 只能是 0，如果超过 0，那么就会生成同样多个零值的实例
	// 或者用`var list []UserLoginListResponse`声明方式也可以
	//var list = make([]UserLoginListResponse, 0)
	var list []UserLoginListResponse
	for _, logModel := range _list {
		list = append(list, UserLoginListResponse{
			ID:           logModel.ID,
			CreatedAt:    logModel.CreatedAt,
			Title:        logModel.Title,
			IP:           logModel.IP,
			IPLocation:   logModel.IPLocation,
			LoginType:    logModel.LoginType,
			UA:           logModel.UA,
			UserID:       logModel.UserID,
			Username:     logModel.UserModel.Username,
			UserNickName: logModel.UserModel.Nickname,
		})
	}

	res.SuccessWithList(list, count, c)
}
