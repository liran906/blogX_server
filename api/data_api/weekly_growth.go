// Path: ./api/data_api/weekly_growth.go

package data_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/redis_service/redis_site"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

type WeeklyGrowthDataReq struct {
	Type int8 `form:"type" binding:"required,oneof=1 2 3 4"` // 1-流量 2-点击率 3-新增文章 4-新增用户
}

type WeeklyGrowthDataResp struct {
	GrowthRate  string         `json:"growthRate"`
	GrowthValue int            `json:"growthValue"`
	ResultMap   map[string]int `json:"resultMap"`
}

type Table struct {
	Date  string `gorm:"column:date"`
	Count int    `gorm:"column:count"`
}

func (DataApi) WeeklyGrowthDataView(c *gin.Context) {
	req := c.MustGet("bindReq").(WeeklyGrowthDataReq)

	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)
	var dataList []Table
	var cachedCount int

	switch req.Type {
	case 1: // 1-流量
		global.DB.Model(models.DataModel{}).Where("date >= ? and date <= ?",
			weekAgo.Format("2006-01-02")+" 00:00:00",
			now.Format("2006-01-02 15:04:05")).
			Select("date", "flow_count AS count").
			Scan(&dataList)
		cachedCount = redis_site.GetFlow(now.Format("2006-01-02")) // 记录 redis 数据
	case 2: // 2-点击率
		global.DB.Model(models.DataModel{}).Where("date >= ? and date <= ?",
			weekAgo.Format("2006-01-02")+" 00:00:00",
			now.Format("2006-01-02 15:04:05")).
			Select("date", "click_count AS count").
			Scan(&dataList)
		cachedCount = redis_site.GetClick(now.Format("2006-01-02")) // 记录 redis 数据
	case 3: // 3-新增文章
		global.DB.Model(models.ArticleModel{}).Where("created_at >= ? and created_at <= ? and status = ?",
			weekAgo.Format("2006-01-02")+" 00:00:00",
			now.Format("2006-01-02 15:04:05"),
			enum.ArticleStatusPublish).
			Select("date(created_at) AS date", "count(id) AS count").
			Group("date").Scan(&dataList)
	case 4: // 4-新增用户
		global.DB.Model(models.UserModel{}).Where("created_at >= ? and created_at <= ?",
			weekAgo.Format("2006-01-02")+" 00:00:00",
			now.Format("2006-01-02 15:04:05")).
			Select("date(created_at) as date", "count(id) as count").
			Group("date").Scan(&dataList)
	}
	var dateMap = map[string]int{}
	for _, data := range dataList {
		date := strings.Split(data.Date, "T")[0]
		dateMap[date] = data.Count
	}

	resp := WeeklyGrowthDataResp{}
	resp.ResultMap = map[string]int{}
	var sixth string
	var last string
	for i := 0; i < 7; i++ {
		dateStr := weekAgo.AddDate(0, 0, i+1).Format("2006-01-02")
		count, _ := dateMap[dateStr]
		resp.ResultMap[dateStr] = count
		if i == 5 {
			sixth = dateStr
		}
		if i == 6 {
			last = dateStr
			resp.ResultMap[dateStr] += cachedCount // 加上 redis 数据
		}
	}

	// 算增长，找最后一个和最后一个的前一个
	resp.GrowthValue = resp.ResultMap[last] - resp.ResultMap[sixth]
	if resp.ResultMap[sixth] == 0 {
		if resp.ResultMap[last] == 0 {
			resp.GrowthRate = "0%"
		} else {
			resp.GrowthRate = "infinite%"
		}
	} else {
		resp.GrowthRate = fmt.Sprintf("%.1f%%", float64(resp.GrowthValue)/float64(resp.ResultMap[sixth])*100)
	}
	res.SuccessWithData(resp, c)
}
